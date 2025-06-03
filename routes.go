package main

import (
	"forum/authentication"
	"forum/post"
	"forum/utils"
	"net/http"
)

func registerRoutes() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/error", error404Page)
	http.HandleFunc("/register", authentication.RegisterHandler)
	http.HandleFunc("/logout", authentication.LogoutHandler)
	http.HandleFunc("/login", authentication.LoginHandler)
	http.HandleFunc("/newpost", utils.RequireAuth(post.NewPost))
	http.HandleFunc("/myposts", utils.RequireAuth(myPostsHandler))
	http.HandleFunc("/newcomment", utils.RequireAuth(post.NewComment))
	http.HandleFunc("/post", handleViewPost)
	http.HandleFunc("/posts/like", (postLikeHandler))
	http.HandleFunc("/comments/like", commentLikeHandler)
	http.HandleFunc("/post/delete", utils.RequireAuth(deletePostHandler))
	http.HandleFunc("/post/edit", utils.RequireAuth(EditPostHandler))
	http.HandleFunc("/comment/delete", utils.RequireAuth(DeleteCommentHandler))
	http.HandleFunc("/comment/edit", utils.RequireAuth(EditCommentHandler))

	// OAuth routes for Google and Facebook
	http.HandleFunc("/auth/google", authentication.GoogleAuthHandler)
	http.HandleFunc("/auth/google/callback", authentication.GoogleCallbackHandler)
	http.HandleFunc("/auth/facebook", authentication.FacebookAuthHandler)
	http.HandleFunc("/auth/facebook/callback", authentication.FacebookCallbackHandler)
	http.HandleFunc("/auth/github", authentication.GitHubAuthHandler)
	http.HandleFunc("/auth/github/callback", authentication.GitHubCallbackHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	//notifications
	http.HandleFunc("/notifications", notificationsPageHandler)
	http.HandleFunc("/notifications/count", notificationsCountHandler)
	http.HandleFunc("/notifications/mark_read", notificationsMarkReadHandler)

}
