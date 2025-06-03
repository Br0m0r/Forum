package post

import (
	"database/sql"
	"fmt"
	"forum/db"
	"forum/likes"
	"forum/notifications"
	"forum/utils"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"net/http"
	"strconv"
)

func NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(20 << 20) // 20MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["categories"]

	if len(title) < 5 || len(title) > 100 {
		http.Error(w, "Title must be between 5 and 100 characters", http.StatusBadRequest)
		return
	}

	if len(content) < 10 || len(content) > 1000 {
		http.Error(w, "Content must be between 10 and 1000 characters", http.StatusBadRequest)
		return
	}

	if len(categories) == 0 {
		http.Error(w, "At least one category must be selected", http.StatusBadRequest)
		return
	}

	cookie := utils.CheckCookie(r)
	userID := utils.GetUserID(cookie)
	username := utils.GetUserName(cookie)
	if userID <= 0 || username == "" {
		http.Error(w, "Unauthorized: invalid session", http.StatusUnauthorized)
		return
	}

	var imagePath string
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/gif":  true,
		}
		if !allowedTypes[handler.Header.Get("Content-Type")] {
			http.Error(w, "Unsupported file type", http.StatusBadRequest)
			return
		}

		fileExt := filepath.Ext(handler.Filename)
		imageFileName := fmt.Sprintf("post_img_%d%s", time.Now().UnixNano(), fileExt)
		imagePath = "post_img/" + imageFileName

		dst, err := os.Create("./static/" + imagePath)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
	}

	result, err := db.Database.Exec(`
        INSERT INTO posts (title, content, user_id, user_name, image) 
        VALUES (?, ?, ?, ?, ?)`, title, content, userID, username, imagePath)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	postID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve post ID", http.StatusInternalServerError)
		return
	}

	for _, cat := range categories {
		catID, err := strconv.Atoi(cat)
		if err != nil {
			continue
		}
		_, err = db.Database.Exec(`
            INSERT INTO post_categories (post_id, category_id) 
            VALUES (?, ?)`, postID, catID)
		if err != nil {
			log.Printf("Failed to insert category: %v", err)
			continue
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GetCategories() []utils.Category {
	rows, err := db.Database.Query("SELECT id, name FROM categories")
	if err != nil {
		fmt.Println("Error fetching categories:", err)
		return nil
	}
	defer rows.Close()

	var categories []utils.Category
	for rows.Next() {
		var c utils.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			continue
		}
		categories = append(categories, c)
	}
	return categories
}

func GetPostByID(id int) utils.Post {
	var post utils.Post
	var img sql.NullString // to handle NULL image values

	// Fetch core post data
	query := `
		SELECT id, title, content, user_id, user_name, image, created_at
		FROM posts 
		WHERE id = ? 
	`

	err := db.Database.QueryRow(query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&post.UserName,
		&img,
		&post.CreatedAt,
	)

	if err != nil {
		fmt.Println("Error retrieving post:", err)
		return utils.Post{}
	}

	// Handle NULL image
	if img.Valid {
		post.Image = img.String
	}
	log.Println("Post image path:", post.Image)

	// Count likes and dislikes
	likesCount, dislikesCount, _ := likes.CountLikes(post.ID)
	post.Likes_count = likesCount
	post.Dislikes_count = dislikesCount

	// Fetch categories associated with this post
	categoryQuery := `
		SELECT c.id, c.name
		FROM categories c
		INNER JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
	`

	rows, err := db.Database.Query(categoryQuery, post.ID)
	if err != nil {
		fmt.Println("Error retrieving categories for post:", err)
		return post
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err == nil {
			post.CategoryIDs = append(post.CategoryIDs, id)
			post.CategoryNames = append(post.CategoryNames, name)
		}
	}

	return post
}

func GetPosts() []utils.Post {
	// First, fetch core post fields
	var img sql.NullString
	rows, err := db.Database.Query(`SELECT id, title, content, user_id, user_name, created_at,image FROM posts ORDER BY created_at DESC`)
	if err != nil {
		fmt.Println("Error retrieving posts:", err)
		return []utils.Post{}
	}
	defer rows.Close()

	var posts []utils.Post

	for rows.Next() {
		var p utils.Post
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Content, &p.UserID, &p.UserName, &p.CreatedAt, &img,
		); err != nil {
			log.Println("12Error scanning post row:", err)
			continue
		}
		if img.Valid {
			p.Image = img.String
		}

		// Likes/Dislikes
		p.Likes_count, p.Dislikes_count, _ = likes.CountLikes(p.ID)

		// Fetch all categories for this post
		catRows, err := db.Database.Query(`
			SELECT c.id, c.name
			FROM categories c
			INNER JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?
		`, p.ID)

		if err != nil {
			log.Println("Error retrieving categories for post ID", p.ID, ":", err)
			continue
		}

		for catRows.Next() {
			var catID int64
			var catName string
			if err := catRows.Scan(&catID, &catName); err == nil {
				p.CategoryIDs = append(p.CategoryIDs, catID)
				p.CategoryNames = append(p.CategoryNames, catName)
			}
		}
		catRows.Close()

		posts = append(posts, p)
	}

	return posts
}

func MyPosts(w http.ResponseWriter, r *http.Request) utils.GroupedPosts {
	cookie := utils.CheckCookie(r)
	userID := utils.GetUserID(cookie)

	var img sql.NullString

	rows, err := db.Database.Query(
		`SELECT id, title, content, user_name, created_at, image,'authored' as post_type
    FROM posts
    WHERE user_id = ?

    UNION

    SELECT p.id, p.title, p.content, p.user_name, p.created_at,image, 'reacted' as post_type
    FROM posts p
    JOIN post_likes pp ON p.id = pp.post_id
    WHERE pp.user_id = ? ORDER BY created_at DESC`, userID, userID)

	if err != nil {
		log.Println("Error retrieving my posts:", err)
		return utils.GroupedPosts{}
	}
	defer rows.Close()

	var authored []utils.Post
	var reacted []utils.Post

	for rows.Next() {
		var p utils.Post
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Content, &p.UserName, &p.CreatedAt, &img, &p.PostType,
		); err != nil {
			log.Println("11Error scanning post row:", err)
			continue
		}

		if img.Valid {
			p.Image = img.String
		}

		// Likes/Dislikes
		p.Likes_count, p.Dislikes_count, _ = likes.CountLikes(p.ID)

		// Fetch all categories for this post
		catRows, err := db.Database.Query(`
			SELECT c.id, c.name
			FROM categories c
			INNER JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?
		`, p.ID)

		if err != nil {
			log.Println("Error retrieving categories for post ID", p.ID, ":", err)
			continue
		}

		for catRows.Next() {
			var catID int64
			var catName string
			if err := catRows.Scan(&catID, &catName); err == nil {
				p.CategoryIDs = append(p.CategoryIDs, catID)
				p.CategoryNames = append(p.CategoryNames, catName)
			}
		}
		catRows.Close()
		// group by type
		if p.PostType == "authored" {
			authored = append(authored, p)
		} else if p.PostType == "reacted" {
			reacted = append(reacted, p)
		}
	}

	return utils.GroupedPosts{
		Authored: authored,
		Reacted:  reacted,
	}
}

func GetCommentsByPostID(postID int) []utils.Comment {
	// no more c.likes_count / c.dislikes_count in SELECT
	const query = `
        SELECT
            c.id,
            c.content,
            c.post_id,
            c.user_id,
            u.username,
            c.created_at
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC
    `
	rows, err := db.Database.Query(query, postID)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return nil
	}
	defer rows.Close()

	var comments []utils.Comment
	for rows.Next() {
		var cm utils.Comment
		if err := rows.Scan(
			&cm.ID,
			&cm.Content,
			&cm.PostID,
			&cm.UserID,
			&cm.UserName,
			&cm.CreatedAt,
		); err != nil {
			log.Printf("Error scanning comment: %v", err)
			continue
		}
		// fetch live counts per comment
		cm.Likes_count, cm.Dislikes_count, _ = likes.CountCommentLikes(cm.ID)
		comments = append(comments, cm)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
	}
	return comments
}

func NewComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

	// Get form data
	content := r.FormValue("content")
	postIDStr := r.FormValue("post_id")

	// Validate form data
	if len(content) < 5 || len(content) > 500 {
		http.Error(w, "Comment must be between 5 and 500 characters", http.StatusBadRequest)
		return
	}

	if postIDStr == "" {
		http.Error(w, "Missing Post ID", http.StatusBadRequest)
		return
	}

	// Convert postID to integer
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	// Get user session
	userCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	cookieID := userCookie.Value

	// Find user ID from session
	var userID int
	err = db.Database.QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookieID).Scan(&userID)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Insert comment into database
	res, err := db.Database.Exec(`
		INSERT INTO comments (content, post_id, user_id, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, content, postID, userID)

	if err != nil {
		http.Error(w, "Failed to insert comment", http.StatusInternalServerError)
		return
	}

	// Send notification to post author (if different from commenter)
	newCommentID, _ := res.LastInsertId()
	var postOwnerID int64
	err = db.Database.QueryRow(
		`SELECT user_id FROM posts WHERE id = ?`, postID,
	).Scan(&postOwnerID)
	if err == nil && postOwnerID != int64(userID) {
		cid := newCommentID
		_ = notifications.Create(postOwnerID, int64(userID), int64(postID), &cid, "comment")
	}

	// Redirect to the post page
	http.Redirect(w, r, fmt.Sprintf("/post?id=%d", postID), http.StatusSeeOther)
}

// CommentedPosts returns each post the user has commented on,
// with all of that user's comments in the Post.Comment slice.
func CommentedPosts(userID int64) ([]utils.Post, error) {
	const sqlStmt = `
    SELECT
      p.id, p.title, p.content, p.user_name, p.created_at, p.image,
      c.id, c.content, c.user_id, c.created_at
    FROM posts p
    JOIN comments c ON p.id = c.post_id
    WHERE c.user_id = ?
    ORDER BY p.created_at DESC, c.created_at ASC
    `
	rows, err := db.Database.Query(sqlStmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map from postID to *Post so we can append comments
	postsMap := make(map[int64]*utils.Post)

	for rows.Next() {
		var (
			pid         int64
			title       string
			content     string
			userName    string
			postCreated time.Time
			imgNull     sql.NullString

			commentID      int64
			commentContent string
			commentUserID  int64
			commentCreated time.Time
		)

		if err := rows.Scan(
			&pid, &title, &content, &userName, &postCreated, &imgNull,
			&commentID, &commentContent, &commentUserID, &commentCreated,
		); err != nil {
			log.Println("scan commented post:", err)
			continue
		}

		// Initialize the post in map if first time
		p, exists := postsMap[pid]
		if !exists {
			p = &utils.Post{
				ID:        pid,
				Title:     title,
				Content:   content,
				UserName:  userName,
				CreatedAt: postCreated,
			}
			if imgNull.Valid {
				p.Image = imgNull.String
			}
			postsMap[pid] = p
		}

		// Append this comment
		p.Comment = append(p.Comment, utils.Comment{
			ID:        commentID,
			Content:   commentContent,
			UserID:    commentUserID,
			CreatedAt: commentCreated,
		})
	}

	// Flatten map into slice
	var result []utils.Post
	for _, p := range postsMap {
		result = append(result, *p)
	}
	return result, nil
}

// GetCommentByID fetches a single comment (for ownership checks).
func GetCommentByID(id int) utils.Comment {
	var c utils.Comment
	err := db.Database.QueryRow(
		"SELECT id, content, post_id, user_id, created_at FROM comments WHERE id = ?",
		id,
	).Scan(&c.ID, &c.Content, &c.PostID, &c.UserID, &c.CreatedAt)
	if err != nil {
		return utils.Comment{}
	}
	// (You can also populate c.UserName, Likes_count, etc., if needed.)
	return c
}
