package authentication

import (
	"fmt"
	"forum/utils"
	"net/http"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// when user first opens the reg page, method is GET:
	if r.Method == http.MethodGet {
		utils.RenderTemplate(w, "register.html", nil)
		return
	}

	// when user has filled the required fields and tries to submit, method is POST so:
	username := r.FormValue("username")
	email := r.FormValue("email")

	if len(username) < 3 || len(username) > 15 {
		utils.RenderTemplate(w, "register.html", "Username must be between 3 and 15 characters long.")
		return
	}
	if !isEmailValid(email) {
		utils.RenderTemplate(w, "register.html", "Invalid email address")
		return
	}
	password := r.FormValue("password")
	if !isPassValid(password) {
		utils.RenderTemplate(w, "register.html", "Password needs to be 8 -15 characters long and to contain at least one number")
		return
	}
	if username == "" || email == "" || password == "" {
		utils.RenderTemplate(w, "register.html", "All fields are required!")
		return
	}

	// checking if the input username OR email already exists
	emailExists, usernameExists, emailProvider, checkUserExistsError := checkRegCredentials(email, username)

	if checkUserExistsError != nil {
		http.Error(w, "Internal Server Error - Error checking Database, plz try again", http.StatusInternalServerError)
		return
	} else if emailExists && emailProvider != "" {
		utils.RenderTemplate(w, "register.html", fmt.Sprintf("You already have an account using %s. Please log in with that provider.", emailProvider))
		return
	} else if emailExists && usernameExists {
		utils.RenderTemplate(w, "register.html", "Username and email already in use!")
		return
	} else if usernameExists {
		utils.RenderTemplate(w, "register.html", "Username already in use!")
		return
	} else if emailExists {
		utils.RenderTemplate(w, "register.html", "Email already in use!")
		return
	}

	hashedPassword, hashedPassErr := hashPassword(password)
	if hashedPassErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// registering the user in the database
	registerError := registerUser(username, email, hashedPassword)
	if registerError != nil {
		http.Error(w, "Internal Server Error - adding to Database", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
