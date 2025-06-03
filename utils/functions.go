package utils

import (
	"forum/db"
	"html/template"
	"log"
	"net/http"
)

// we read and parse all .html files in the templates directory
// and we combine them into a single *template.Template object stored in tmpl variable
// if we dont want to parse/load all templates in advance, like in larger projects,
// we can just load each time only the one we need (via a custom function):
var Tmpl = template.Must(template.ParseGlob("static/templates/*.html"))

// the third argument is the data we provide for the template execution
// it corresponds to the {{.}} part of the .html file
func RenderTemplate(w http.ResponseWriter, name string, data any) {
	err := Tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "HTTP status 500 - Could not render template", http.StatusInternalServerError)
		return
	}
}

// middleware <<see routes.go>>
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
		return -1 // or 0 if you prefer, but -1 is clearer for "not found"
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
	userRowErr := userRow.Scan(&username)
	if userRowErr != nil {
		return ""
	}
	if username != "" {
		return username
	} else {
		return ""
	}
}
