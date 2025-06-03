package authentication

import (
	"database/sql"
	"fmt"
	"forum/db"
)

// this is Step 4 helper function:
func findOrCreateUser(email, name, provider string) (int, error) {
	var userID int
	var existingProvider sql.NullString

	userRow := db.Database.QueryRow("SELECT id FROM users WHERE email = ?", email)
	// QueryRow() is a method of *sql.DB. It executes the SQL query on our "database" and returns a single row
	userRowErr := userRow.Scan(&userID)
	// Scan() is a method of *sql.Row. It assigns the value of the row to a variable (userID here) OR
	// returns an error if something goes wrong

	if userRowErr == sql.ErrNoRows {
		// no row returned - user doesnt exist in our DB
		insertUser, insertUserErr := db.Database.Exec("INSERT INTO users (username, email, provider) VALUES (?, ?, ?)", name, email, provider)
		// Exec() is a method of *sql.DB that returns an sql.Result (the insertion of the new user in "users" table here) or an error
		if insertUserErr != nil {
			return 0, fmt.Errorf("failed to insert new user: %w", insertUserErr)
		}
		lastID, err := insertUser.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID: %w", err)
		}

		// this is the user id (primary key of "users" table) of the one we just inserted above
		userID = int(lastID)
		return userID, nil

	} else if userRowErr != nil {
		// some other error
		return 0, fmt.Errorf("failed to check user existence: %w", userRowErr)
	}

	// if provider is null, or different, we update it
	if !existingProvider.Valid || existingProvider.String != provider {
		_, updateErr := db.Database.Exec("UPDATE users SET provider = ? WHERE id = ?", provider, userID)
		if updateErr != nil {
			return 0, fmt.Errorf("failed to update provider: %w", updateErr)
		}
	}

	return userID, nil

}
