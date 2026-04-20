package utils

import (
	"forum/db"
	"html/template"
	"log"
	"net/http"
)

var Tmpl = template.Must(template.ParseGlob("static/templates/*.html"))
func RenderTemplate(w http.ResponseWriter, name string, data any) {
	err := Tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "HTTP status 500 - Could not render template", http.StatusInternalServerError)
		return
	}
}

// RequireAuth is middleware that redirects unauthenticated users to the home page.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func GetUserID(cookie string) int {
	var userID int
	err := db.Database.
		QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie).
		Scan(&userID)
	if err != nil {
		log.Printf("GetUserID error: %v", err)
		return -1
	}
	return userID
}

func CheckCookie(r *http.Request) string {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("Error retrieving session_token cookie: %v", err)

		return ""
	}
	return cookie.Value
}

func GetUserName(cookie string) string {

	var username string

	userRow := db.Database.QueryRow("SELECT users.username FROM users INNER JOIN sessions ON users.id = sessions.user_id WHERE sessions.id = ?", cookie)
	if err := userRow.Scan(&username); err != nil {
		return ""
	}
	return username
}
