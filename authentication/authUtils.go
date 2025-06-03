package authentication

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/db"
	"net/http"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GENERATING AND STORING A NEW SESSION ID (using UUID) AND A NEW COOKIE
func createSession(userID int, w http.ResponseWriter) {
	// deleting any existing session:
	_, err := db.Database.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to remove old session", http.StatusInternalServerError)
		return
	}

	// generating a new Session ID and storing it in DB:
	sessionID := uuid.New().String()
	expirationTime := time.Now().Add(24 * time.Hour) // 1-day session

	_, err = db.Database.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expirationTime)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// creating a cookie
	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expirationTime,
		HttpOnly: true, // we use this to prevent JavaScript access
		Path:     "/",  // this way the cookie is available across the entire site
	}

	// setting the cookie in the user's browser
	http.SetCookie(w, &cookie)
}

// HASHING THE USER PASSWORD PROVIDED BY THE USER BEFORE STORING IN DATABASE
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // DefaultCost = 10
	return string(hashed), err
}

// CHECKING IF PASSWORD INPUT IS VALID
func isPassValid(pass string) bool {
	if len(pass) < 8 || len(pass) > 15 {
		return false
	}
	for _, letter := range pass {
		if letter >= '0' && letter <= '9' {
			return true
		}
	}
	return false
}

// CHECKING IF EMAIL INPUT IS VALID
func isEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	// returns the email parsed or an error if the email argument wasnt in correct format
	return err == nil
	// so if error is nil (the ParseAddress didnt return an error) this returns True,
	// because the statement err == nil IS True, otherwise this statement is False and
	// isEmailValid returns False then

	// despite we use type = "email" in our .html file, it is good to also check it in the backend
}

// CHECKING IN DATASBASE IF A USER WITH THE SAME EMAIL AND/OR USERNAME ALREADY EXISTS (while registering a new user)
func checkRegCredentials(email, username string) (bool, bool, string, error) {
	var emailExists, usernameExists bool
	var provider sql.NullString

	emailRow := db.Database.QueryRow("SELECT provider FROM users WHERE email = ?", email)
	emailErr := emailRow.Scan(&provider)
	if emailErr != nil && emailErr != sql.ErrNoRows {
		// when Scan(&variable) doesn't find any matching row, it returns the error: sql.ErrNoRows
		// so to return a real error (like a database connection error) we need to exclude the "no rows found" error
		return false, false, "", emailErr
	}

	var emailProvider string

	if emailErr == nil {
		emailExists = true
		if provider.Valid {
			emailProvider = provider.String
		}
	}

	usernameRow := db.Database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username)
	// QueryRow() is a method of *sql.DB. It executes the SQL query on our "database" and returns a single row
	// bc of "EXISTS" in the query, the result is either 1 or 0
	usernameErr := usernameRow.Scan(&usernameExists)
	// Scan() is a method of *sql.Row. It assigns the value of the row to a variable (an "exists") OR
	// returns an error if something goes wrong. So here exists is either 1 or 0 (T or F)
	if usernameErr != nil && usernameErr != sql.ErrNoRows {
		return false, false, "", usernameErr
	}

	return emailExists, usernameExists, emailProvider, nil
}

// CHECKING IN DATABASE IF THE INPUT EMAIL AND PASSWORD ARE CORRECT (while logging in an existing user)
func checkLoginCredentials(email, password string) (bool, error) {
	var storedHashedPassword sql.NullString

	// getting the hashed password from the database
	getHashedPassword := db.Database.QueryRow("SELECT password FROM users WHERE email = ?", email)
	// from this SQL query we extract the value from the password column and with the next line
	// we assign it to the variable getHashedPassword
	getHashedPasswordErr := getHashedPassword.Scan(&storedHashedPassword)
	if getHashedPasswordErr != nil {
		if getHashedPasswordErr == sql.ErrNoRows { // no matching email found in the database
			return false, nil
		}
		return false, getHashedPasswordErr // other database errors
	}

	// this is for OAuth users (password is NULL)
	if !storedHashedPassword.Valid || storedHashedPassword.String == "" {
		return false, ErrOAuthUser
	}
	// comparing the hashed password we got before, with the provided password
	compareErr := bcrypt.CompareHashAndPassword([]byte(storedHashedPassword.String), []byte(password))
	if compareErr != nil {
		return false, nil
	}

	// if we reach this return, it means everything is correct, so we login the user
	return true, nil
}

// REGISTERING THE NEW USER IN DATABASE
func registerUser(username, email, hashedPassword string) error {
	query, err := db.Database.Prepare("INSERT INTO users (username, email, password) VALUES (?, ?, ?)")
	// Prepare() is a method of *sql.DB. It creates a statement for later use (usually with Exe())
	if err != nil {
		return err
	}
	defer query.Close()

	_, err = query.Exec(username, email, hashedPassword)
	// Exec() is a method of *sql.Stmt that executes the query/statement with the given arguments (username,
	// email, pass) or an error if something goes wrong
	return err
}

// GETTING THE USER'S ID FROM EMAIL
func getUserID(email string) (int, error) {
	// getting the user id of the email provided:
	// (user id is the primary key - autoincremented first column of users table)
	var userID int
	userRow := db.Database.QueryRow("SELECT id FROM users WHERE email = ?", email)
	userRowErr := userRow.Scan(&userID)
	// the two lines above can be written in one, we used this syntax to be more clear
	if userRowErr != nil {
		return 0, fmt.Errorf("error fetching user ID: %w", userRowErr)
	}
	return userID, nil
}

// SPECIFIC ERROR WHEN AN OAUTH-REGISTERED USER TRIES TO LOGIN USING A PASSWORD
var ErrOAuthUser = errors.New("oAuth user, cannot use password login")
