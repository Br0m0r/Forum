package main

import (
	"encoding/json"
	"fmt"
	"forum/db"
	"forum/likes"
	"forum/notifications"
	"forum/post"
	"forum/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// homeHandler shows the list of posts, optionally filtered by category.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		error404Page(w, r)
		return
	}

	categoryIDStr := r.URL.Query().Get("category")
	var posts []utils.Post
	var selectedCategory int64
	var err error

	if categoryIDStr != "" {
		categoryID, parseErr := strconv.ParseInt(categoryIDStr, 10, 64)
		if parseErr != nil {
			log.Println("Invalid category ID:", parseErr)
			posts = post.GetPosts()
		} else {
			posts, err = post.FilteredPosts(categoryID)
			if err != nil {
				log.Println("Error filtering posts:", err)
				posts = post.GetPosts()
			}
			selectedCategory = categoryID
		}
	} else {
		posts = post.GetPosts()
	}

	// Populate UserVote for each post
	token := utils.CheckCookie(r)
	userID := utils.GetUserID(token)
	if userID > 0 {
		for i := range posts {
			posts[i].UserVote, _ = likes.GetUserPostVote(int64(userID), posts[i].ID)
		}
	}

	data := utils.TemplateData{
		Username:         utils.GetUserName(token),
		Posts:            posts,
		Categories:       post.GetCategories(),
		SelectedCategory: selectedCategory,
	}
	utils.RenderTemplate(w, "home.html", data)
}

// handleViewPost shows a single post and its comments.
func handleViewPost(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post" {
		error404Page(w, r)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	token := utils.CheckCookie(r)
	userID := utils.GetUserID(token)

	data := utils.TemplateData{
		Username: utils.GetUserName(token),
		Post:     post.GetPostByID(id),
		Comments: post.GetCommentsByPostID(id),
	}

	if userID > 0 {
		data.Post.UserVote, _ = likes.GetUserPostVote(int64(userID), data.Post.ID)
		for i := range data.Comments {
			data.Comments[i].UserVote, _ = likes.GetUserCommentVote(int64(userID), data.Comments[i].ID)
		}
	}

	if data.Post.ID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	utils.RenderTemplate(w, "post.html", data)
}

// myPostsHandler shows posts the logged-in user authored, liked, or commented on.
func myPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/myposts" {
		error404Page(w, r)
		return
	}

	token := utils.CheckCookie(r)
	userID := utils.GetUserID(token)

	// 1) Authored & Reacted
	grouped := post.MyPosts(w, r)

	// 2) Commented posts (each post once with all comments)
	commentedPosts, err := post.CommentedPosts(int64(userID))
	if err != nil {
		log.Println("Error fetching commented posts:", err)
		commentedPosts = []utils.Post{}
	}

	// 3) Populate counts and UserVote for each list
	if userID > 0 {
		// Authored
		for i := range grouped.Authored {
			grouped.Authored[i].UserVote, _ = likes.GetUserPostVote(int64(userID), grouped.Authored[i].ID)
		}
		// Reacted
		for i := range grouped.Reacted {
			grouped.Reacted[i].UserVote, _ = likes.GetUserPostVote(int64(userID), grouped.Reacted[i].ID)
		}
		// Commented: first load counts, then vote state
		for i := range commentedPosts {
			l, d, _ := likes.CountLikes(commentedPosts[i].ID)
			commentedPosts[i].Likes_count = l
			commentedPosts[i].Dislikes_count = d
			commentedPosts[i].UserVote, _ = likes.GetUserPostVote(int64(userID), commentedPosts[i].ID)
		}
	}

	// 4) Render template, reusing .Posts for commented posts
	data := utils.TemplateData{
		Username:   utils.GetUserName(token),
		Categories: post.GetCategories(),
		Authored:   grouped.Authored,
		Reacted:    grouped.Reacted,
		Posts:      commentedPosts,
	}
	utils.RenderTemplate(w, "myposts.html", data)
}

func postLikeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("postLikeHandler: Started with URL=%s, Method=%s", r.URL.String(), r.Method)

	if r.Method != http.MethodPost {
		log.Printf("postLikeHandler: Invalid method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("postLikeHandler: No session cookie found: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int64
	if err := db.Database.
		QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID); err != nil {
		log.Printf("postLikeHandler: Invalid session %q: %v", cookie.Value, err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("postLikeHandler: Authenticated user_id=%d", userID)

	pidStr := r.URL.Query().Get("post_id")
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		log.Printf("postLikeHandler: Invalid post_id=%q: %v", pidStr, err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	log.Printf("postLikeHandler: Processing like for post_id=%d", pid)

	// Get the current vote status
	var currentVote int
	err = db.Database.QueryRow(
		"SELECT CASE WHEN EXISTS(SELECT 1 FROM post_votes WHERE user_id = ? AND post_id = ? AND is_like = 1) THEN 1 "+
			"WHEN EXISTS(SELECT 1 FROM post_votes WHERE user_id = ? AND post_id = ? AND is_like = 0) THEN -1 "+
			"ELSE 0 END",
		userID, pid, userID, pid,
	).Scan(&currentVote)
	if err != nil {
		log.Printf("postLikeHandler: Error checking current vote: %v", err)
		currentVote = 0 // Assume no vote if error
	}
	log.Printf("postLikeHandler: Current vote status=%d for user_id=%d, post_id=%d",
		currentVote, userID, pid)

	isLike := r.URL.Query().Get("is_like") == "1"
	log.Printf("postLikeHandler: Vote request is_like=%v", isLike)

	// Determine if we should notify
	shouldNotify := false
	if currentVote == 0 || // No previous vote (adding new like/dislike)
		(currentVote == 1 && !isLike) || // Changing from like to dislike
		(currentVote == -1 && isLike) { // Changing from dislike to like
		shouldNotify = true
		log.Printf("postLikeHandler: Should send notification = true (adding or changing vote)")
	} else if (currentVote == 1 && isLike) || (currentVote == -1 && !isLike) {
		// Don't notify when removing a like/dislike
		shouldNotify = false
		log.Printf("postLikeHandler: Should send notification = false (removing vote)")
	} else {
		log.Printf("postLikeHandler: Should send notification = false (other case)")
	}

	// Toggle the like/dislike
	if err := likes.ToggleLike(userID, pid, isLike); err != nil {
		log.Printf("postLikeHandler: ToggleLike error: %v", err)
		http.Error(w, "Failed to register vote", http.StatusInternalServerError)
		return
	}
	log.Printf("postLikeHandler: Successfully toggled vote for user_id=%d, post_id=%d", userID, pid)

	// Create notification if needed
	if shouldNotify {
		var postAuthorID int64
		if err := db.Database.QueryRow("SELECT user_id FROM posts WHERE id = ?", pid).Scan(&postAuthorID); err != nil {
			log.Printf("postLikeHandler: Error getting post author: %v", err)
		} else if postAuthorID != userID { // Don't notify users of their own likes
			notifKind := "like"
			if !isLike {
				notifKind = "dislike"
			}
			log.Printf("postLikeHandler: Creating %s notification for post_author_id=%d from user_id=%d",
				notifKind, postAuthorID, userID)
			if err := notifications.Create(postAuthorID, userID, pid, nil, notifKind); err != nil {
				log.Printf("postLikeHandler: Error creating notification: %v", err)
			}
		} else {
			log.Printf("postLikeHandler: Skipping notification as user liked their own post")
		}
	}

	// Return updated counts
	l, d, _ := likes.CountLikes(pid)
	resp := map[string]int{"Likes_count": l, "Dislikes_count": d}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Printf("postLikeHandler: Returning updated counts: likes=%d, dislikes=%d", l, d)
}

func commentLikeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("commentLikeHandler: Started with URL=%s, Method=%s", r.URL.String(), r.Method)

	if r.Method != http.MethodPost {
		log.Printf("commentLikeHandler: Invalid method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("commentLikeHandler: No session cookie found: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int64
	if err := db.Database.
		QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID); err != nil {
		log.Printf("commentLikeHandler: Invalid session %q: %v", cookie.Value, err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("commentLikeHandler: Authenticated user_id=%d", userID)

	cidStr := r.URL.Query().Get("comment_id")
	cid, err := strconv.ParseInt(cidStr, 10, 64)
	if err != nil {
		log.Printf("commentLikeHandler: Invalid comment_id=%q: %v", cidStr, err)
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	log.Printf("commentLikeHandler: Processing like for comment_id=%d", cid)

	// Ensure comment exists
	var exists bool
	if err := db.Database.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)",
		cid,
	).Scan(&exists); err != nil {
		log.Printf("commentLikeHandler: Error checking comment existence: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		log.Printf("commentLikeHandler: Comment with id=%d not found", cid)
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Get the current vote status
	var currentVote int
	err = db.Database.QueryRow(
		"SELECT CASE WHEN EXISTS(SELECT 1 FROM comment_votes WHERE user_id = ? AND comment_id = ? AND is_like = 1) THEN 1 "+
			"WHEN EXISTS(SELECT 1 FROM comment_votes WHERE user_id = ? AND comment_id = ? AND is_like = 0) THEN -1 "+
			"ELSE 0 END",
		userID, cid, userID, cid,
	).Scan(&currentVote)
	if err != nil {
		log.Printf("commentLikeHandler: Error checking current vote: %v", err)
		currentVote = 0 // Assume no vote if error
	}
	log.Printf("commentLikeHandler: Current vote status=%d for user_id=%d, comment_id=%d",
		currentVote, userID, cid)

	isLike := r.URL.Query().Get("is_like") == "1"
	log.Printf("commentLikeHandler: Vote request is_like=%v", isLike)

	// Determine if we should notify
	shouldNotify := false
	if currentVote == 0 || // No previous vote (adding new like/dislike)
		(currentVote == 1 && !isLike) || // Changing from like to dislike
		(currentVote == -1 && isLike) { // Changing from dislike to like
		shouldNotify = true
		log.Printf("commentLikeHandler: Should send notification = true (adding or changing vote)")
	} else if (currentVote == 1 && isLike) || (currentVote == -1 && !isLike) {
		// Don't notify when removing a like/dislike
		shouldNotify = false
		log.Printf("commentLikeHandler: Should send notification = false (removing vote)")
	} else {
		log.Printf("commentLikeHandler: Should send notification = false (other case)")
	}

	// Toggle the like/dislike
	if err := likes.ToggleCommentLike(userID, cid, isLike); err != nil {
		log.Printf("commentLikeHandler: ToggleCommentLike error: %v", err)
		http.Error(w, "Failed to register vote", http.StatusInternalServerError)
		return
	}
	log.Printf("commentLikeHandler: Successfully toggled vote for user_id=%d, comment_id=%d", userID, cid)

	// Create notification if needed
	if shouldNotify {
		var commentAuthorID int64
		var postID int64
		if err := db.Database.QueryRow("SELECT user_id, post_id FROM comments WHERE id = ?", cid).Scan(&commentAuthorID, &postID); err != nil {
			log.Printf("commentLikeHandler: Error getting comment author: %v", err)
		} else if commentAuthorID != userID { // Don't notify users of their own likes
			notifKind := "comment_like"
			if !isLike {
				notifKind = "comment_dislike"
			}
			log.Printf("commentLikeHandler: Creating %s notification for comment_author_id=%d from user_id=%d",
				notifKind, commentAuthorID, userID)
			if err := notifications.Create(commentAuthorID, userID, postID, &cid, notifKind); err != nil {
				log.Printf("commentLikeHandler: Error creating notification: %v", err)
			}
		} else {
			log.Printf("commentLikeHandler: Skipping notification as user liked their own comment")
		}
	}

	// Return updated counts
	l, d, _ := likes.CountCommentLikes(cid)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"Likes_count":    l,
		"Dislikes_count": d,
	})
	log.Printf("commentLikeHandler: Returning updated counts: likes=%d, dislikes=%d", l, d)
}

// notificationsPageHandler renders the full notifications list page.
func notificationsPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("notificationsPageHandler: Started with URL=%s, Method=%s", r.URL.String(), r.Method)

	// auth: session → userID
	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("notificationsPageHandler: No session cookie found: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	log.Printf("notificationsPageHandler: Found session cookie: %s", cookie.Value)

	var userID int64
	if err := db.Database.
		QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID); err != nil {
		log.Printf("notificationsPageHandler: Invalid session %q: %v", cookie.Value, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	log.Printf("notificationsPageHandler: Authenticated user_id=%d", userID)

	// fetch mark-read
	log.Printf("notificationsPageHandler: Fetching notifications for user_id=%d", userID)
	notifs, err := notifications.List(userID)
	if err != nil {
		log.Printf("notificationsPageHandler: Error listing notifications: %v", err)
		http.Error(w, "Failed to load notifications", http.StatusInternalServerError)
		return
	}
	log.Printf("notificationsPageHandler: Retrieved %d notifications for user_id=%d", len(notifs), userID)

	log.Printf("notificationsPageHandler: Marking notifications as read for user_id=%d", userID)
	if err := notifications.MarkAllRead(userID); err != nil {
		log.Printf("notificationsPageHandler: Warning - Failed to mark notifications as read: %v", err)
		// Continue despite error
	} else {
		log.Printf("notificationsPageHandler: Successfully marked all notifications as read for user_id=%d", userID)
	}

	// render template
	username := utils.GetUserName(cookie.Value)
	log.Printf("notificationsPageHandler: Rendering template for username=%s", username)
	data := utils.TemplateData{
		Username:      username,
		Notifications: notifs,
		NotifCount:    0, // since we just marked them read
	}
	utils.RenderTemplate(w, "notifications.html", data)
	log.Printf("notificationsPageHandler: Successfully rendered notifications page for user_id=%d", userID)
}

// notificationsCountHandler returns unread count as JSON {"count": N}
func notificationsCountHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("notificationsCountHandler: Started with URL=%s, Method=%s", r.URL.String(), r.Method)

	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("notificationsCountHandler: No session cookie found: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("notificationsCountHandler: Found session cookie: %s", cookie.Value)

	var userID int64
	if err := db.Database.
		QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID); err != nil {
		log.Printf("notificationsCountHandler: Invalid session %q: %v", cookie.Value, err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("notificationsCountHandler: Authenticated user_id=%d", userID)

	log.Printf("notificationsCountHandler: Fetching unread count for user_id=%d", userID)
	cnt, err := notifications.UnreadCount(userID)
	if err != nil {
		log.Printf("notificationsCountHandler: Error fetching unread count: %v", err)
		http.Error(w, "Failed to fetch count", http.StatusInternalServerError)
		return
	}
	log.Printf("notificationsCountHandler: Found %d unread notifications for user_id=%d", cnt, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": cnt})
	log.Printf("notificationsCountHandler: Successfully sent count response for user_id=%d", userID)
}

// notificationsMarkReadHandler marks all notifications read (NoContent response)
func notificationsMarkReadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("notificationsMarkReadHandler: Started with URL=%s, Method=%s", r.URL.String(), r.Method)

	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("notificationsMarkReadHandler: No session cookie found: %v", err)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Printf("notificationsMarkReadHandler: Found session cookie: %s", cookie.Value)

	var userID int64
	if err := db.Database.
		QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID); err != nil {
		log.Printf("notificationsMarkReadHandler: Invalid session %q: %v", cookie.Value, err)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Printf("notificationsMarkReadHandler: Authenticated user_id=%d", userID)

	log.Printf("notificationsMarkReadHandler: Marking notifications as read for user_id=%d", userID)
	if err := notifications.MarkAllRead(userID); err != nil {
		log.Printf("notificationsMarkReadHandler: Error marking notifications as read: %v", err)
	} else {
		log.Printf("notificationsMarkReadHandler: Successfully marked all notifications as read for user_id=%d", userID)
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("notificationsMarkReadHandler: Responded with 204 No Content")
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("deletePostHandler: Started with URL=%s, Method=%s", r.URL.String(), r.Method)

	if r.Method != http.MethodPost {
		log.Printf("deletePostHandler: Invalid method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("deletePostHandler: Failed to parse form: %v", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	log.Printf("deletePostHandler: Form parsed successfully")

	// parse and load
	postIDStr := r.FormValue("post_id")
	log.Printf("deletePostHandler: Received post_id=%q", postIDStr)

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		log.Printf("deletePostHandler: Failed to parse post_id=%q: %v", postIDStr, err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	log.Printf("deletePostHandler: Processing deletion for post_id=%d", postID)

	p := post.GetPostByID(postID)
	if p.ID == 0 {
		log.Printf("deletePostHandler: Post with ID=%d not found", postID)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	log.Printf("deletePostHandler: Found post with ID=%d, author=%d, title=%q", p.ID, p.UserID, p.Title)

	// ownership check
	cookie := utils.CheckCookie(r)
	currentUser := utils.GetUserID(cookie)
	log.Printf("deletePostHandler: Current user=%d, post author=%d", currentUser, p.UserID)

	if p.UserID != int64(currentUser) {
		log.Printf("deletePostHandler: Permission denied for user=%d to delete post=%d (owned by user=%d)",
			currentUser, postID, p.UserID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// delete
	log.Printf("deletePostHandler: Attempting to delete post_id=%d", postID)
	if err := post.DeletePostByID(postID); err != nil {
		log.Printf("deletePostHandler: Failed to delete post_id=%d: %v", postID, err)
		http.Error(w, fmt.Sprintf("Failed to delete post: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("deletePostHandler: Successfully deleted post_id=%d", postID)

	log.Printf("deletePostHandler: Redirecting to /myposts after successful deletion of post_id=%d", postID)

	referer := r.Header.Get("Referer")
	redirectURL := "/myposts" // Default fallback

	// Don't redirect back to delete/edit pages
	if referer != "" && !strings.Contains(referer, "/delete") && !strings.Contains(referer, "/edit") {
		redirectURL = referer
	}

	log.Printf("deletePostHandler: Redirecting to %s after successful deletion of post_id=%d", redirectURL, postID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)

}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// parse
	postID64, err := strconv.ParseInt(r.FormValue("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	newContent := strings.TrimSpace(r.FormValue("new_content"))
	if newContent == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	// load + ownership
	p := post.GetPostByID(int(postID64))
	if p.ID == 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	currentUser := utils.GetUserID(utils.CheckCookie(r))
	if p.UserID != int64(currentUser) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// update
	if err := post.EditPostByID(postID64, newContent); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update post: %v", err), http.StatusInternalServerError)
		return
	}

	referer := r.Header.Get("Referer")
	redirectURL := "/myposts" // Default fallback

	// Redirect to the post itself as a good fallback
	redirectURL = fmt.Sprintf("/post?id=%d", postID64)

	// Use referer if it's not a delete/edit page
	if referer != "" && !strings.Contains(referer, "/delete") && !strings.Contains(referer, "/edit") {
		redirectURL = referer
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func EditCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// parse
	commentID64, err := strconv.ParseInt(r.FormValue("contentID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	newContent := strings.TrimSpace(r.FormValue("content"))
	if newContent == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	// load + ownership
	c := post.GetCommentByID(int(commentID64))
	if c.ID == 0 {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}
	currentUser := utils.GetUserID(utils.CheckCookie(r))
	if c.UserID != int64(currentUser) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// update
	if err := post.EditCommentByID(commentID64, newContent); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update comment: %v", err), http.StatusInternalServerError)
		return
	}

	referer := r.Header.Get("Referer")
	redirectURL := "/myposts" // Default fallback

	// If we have the associated post ID, redirect to the post page
	if c.PostID > 0 {
		redirectURL = fmt.Sprintf("/post?id=%d", c.PostID)
	}

	// Use referer if it's not a delete/edit page
	if referer != "" && !strings.Contains(referer, "/delete") && !strings.Contains(referer, "/edit") {
		redirectURL = referer
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("DeleteCommentHandler: Started with URL=%s, Method=%s", r.URL.String(), r.Method)

	if r.Method != http.MethodPost {
		log.Printf("DeleteCommentHandler: Invalid method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("DeleteCommentHandler: Failed to parse form: %v", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	log.Printf("DeleteCommentHandler: Form parsed successfully")

	commentIDStr := r.FormValue("comment_id")
	log.Printf("DeleteCommentHandler: Received comment_id=%q", commentIDStr)

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		log.Printf("DeleteCommentHandler: Failed to parse comment_id=%q: %v", commentIDStr, err)
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	log.Printf("DeleteCommentHandler: Processing deletion for comment_id=%d", commentID)

	c := post.GetCommentByID(commentID)
	if c.ID == 0 {
		log.Printf("DeleteCommentHandler: Comment with ID=%d not found", commentID)
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}
	log.Printf("DeleteCommentHandler: Found comment with ID=%d, author=%d, post=%d", c.ID, c.UserID, c.PostID)

	cookie := utils.CheckCookie(r)
	currentUser := utils.GetUserID(cookie)
	log.Printf("DeleteCommentHandler: Current user=%d, comment author=%d", currentUser, c.UserID)

	if c.UserID != int64(currentUser) {
		log.Printf("DeleteCommentHandler: Permission denied for user=%d to delete comment=%d (owned by user=%d)",
			currentUser, commentID, c.UserID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	log.Printf("DeleteCommentHandler: Attempting to delete comment_id=%d", commentID)
	if err := post.DeleteCommentByID(commentID); err != nil {
		log.Printf("DeleteCommentHandler: Failed to delete comment_id=%d: %v", commentID, err)
		http.Error(w, fmt.Sprintf("Failed to delete comment: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("DeleteCommentHandler: Successfully deleted comment_id=%d", commentID)

	log.Printf("DeleteCommentHandler: Redirecting to /myposts after successful deletion of comment_id=%d", commentID)

	referer := r.Header.Get("Referer")
	redirectURL := "/myposts" // Default fallback

	// If we have the associated post ID, redirect to the post page
	if c.PostID > 0 {
		redirectURL = fmt.Sprintf("/post?id=%d", c.PostID)
	}

	// Use referer if it's not a delete/edit page
	if referer != "" && !strings.Contains(referer, "/delete") && !strings.Contains(referer, "/edit") {
		redirectURL = referer
	}

	log.Printf("DeleteCommentHandler: Redirecting to %s after successful deletion of comment_id=%d", redirectURL, commentID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
