package authentication

import (
	"forum/db"
	"net/http"
	"time"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// getting the session cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	sessionID := cookie.Value

	// getting the user's id for this sessionID
	var userID int
	userRow := db.Database.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID)
	userRowErr := userRow.Scan(&userID)
	if userRowErr != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// removing the session from the database
	_, err = db.Database.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	// if the user logs in on two different devices, using this query we log him out
	// from all devices instead of only a particular session using sessionID
	// othwerwise we should use:
	// _, err = db.Database.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		http.Error(w, "Error logging out. Please try again.", http.StatusInternalServerError)
		return
	}

	// expiring the session cookie immediately
	expiredCookie := http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Expire the cookie immediately
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &expiredCookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
