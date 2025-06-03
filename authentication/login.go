package authentication

import (
	"forum/utils"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		utils.RenderTemplate(w, "login.html", nil)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		utils.RenderTemplate(w, "login.html", "All fields are required!")
		return
	}

	// checking if the input email AND password exist in the database
	credentialsCorrect, checkLoginCredentialsError := checkLoginCredentials(email, password)

	if checkLoginCredentialsError != nil {
		// oAuth login w/o password case
		if checkLoginCredentialsError == ErrOAuthUser {
			utils.RenderTemplate(w, "login.html", "You signed up with Google/Facebook/GitHub. Please log in using oAuth.")
			return
		}
		http.Error(w, "Internal Server Error - Error checking Database, plz try again", http.StatusInternalServerError)
		return
	}

	if !credentialsCorrect {
		utils.RenderTemplate(w, "login.html", "Email and/or password are not correct. If you don't have an account try to register first!")
		return
	}

	userID, err := getUserID(email)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	createSession(userID, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
