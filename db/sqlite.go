package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	// SQLite driver
	// with the _ in start, we only run the package's init function
	// we do NOT make its exported functions, types, and variables available to use directly
	// we do not have to use the package elsewhere in the project
)

var Database *sql.DB

func InitDB() error {

	// --------------------------------------------------------------------------------------------------
	// checking if the "./db" directory exists, if not we create it

	_, statErr := os.Stat("./db")
	// os.Stat returns info about a file/folder or an error if it doesnt exist
	// we only need the error here, so we can use it in next check

	if os.IsNotExist(statErr) {
		// os.IsNotExist only takes as arguments error values (errors from os.Stat, os.Open etc)
		// it returns True if the error from os.Stat indicates that the file/directory doesn't exist or
		// False for any other error (the file/folder exists or permission issues etc)

		makeDirErr := os.MkdirAll("./db", 0755)
		if makeDirErr != nil {
			return fmt.Errorf("failed to create database directory: %v", makeDirErr)
		}
	}

	// --------------------------------------------------------------------------------------------------
	// opening database connection and storing it in the variable "database" we created in start

	var openErr error

	Database, openErr = sql.Open("sqlite3", "./db/forum.db")
	// sql.Open returns: *sql.DB, error
	if openErr != nil {
		return fmt.Errorf("failed to open database: %v", openErr)
	}

	pingErr := Database.Ping()
	// Ping() is a method of the sql.DB struct (func (db *sql.DB) Ping() error)
	if pingErr != nil {
		return fmt.Errorf("failed to connect to database: %v", pingErr)
	}

	// Exec() - method of sql.DB struct
	// executes any SQL statement that does NOT return rows, so it works for
	// enabling foreign keys and creating tables
	//(by default, foreign key constraints are disabled in SQLite)
	_, foreignKeyErr := Database.Exec("PRAGMA foreign_keys = ON;")
	if foreignKeyErr != nil {
		return fmt.Errorf("failed to enable foreign keys: %v", foreignKeyErr)
	}

	// --------------------------------------------------------------------------------------------------
	// creating the required tables using the custom function we have defined
	createTablesErr := createTables()
	if createTablesErr != nil {
		return fmt.Errorf("failed to create tables: %v", createTablesErr)
	}

	Database.Exec("INSERT INTO categories (name) VALUES (?), (?), (?)", "Health", "Nature", "Sports")
	log.Println("Database initialized successfully")
	return nil
}

func CloseDB() error {
	if Database != nil {
		return Database.Close()
	}
	return nil
}

// ======================================================================================================
// ======================================================================================================

func createTables() error {
	userTable := `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        password TEXT,
        provider TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	categoryTable := `CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );`

	postTable := `CREATE TABLE IF NOT EXISTS posts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    title      TEXT    NOT NULL,
    content    TEXT    NOT NULL,
    user_id    TEXT    NOT NULL,
    user_name  TEXT    NOT NULL,
	image      TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id)   REFERENCES users(id),
    FOREIGN KEY (user_name) REFERENCES users(username)
	);`

	commentTable := `CREATE TABLE IF NOT EXISTS comments (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        content    TEXT    NOT NULL,
        post_id    INTEGER NOT NULL,
        user_id    INTEGER NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (post_id) REFERENCES posts(id),
        FOREIGN KEY (user_id)  REFERENCES users(id)
    );`

	sessionTable := `CREATE TABLE IF NOT EXISTS sessions (
        id         TEXT    PRIMARY KEY,
        user_id    INTEGER NOT NULL,
        expires_at TIMESTAMP NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );`

	commentLikes := `CREATE TABLE IF NOT EXISTS comment_likes (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id    INTEGER NOT NULL,
        comment_id INTEGER NOT NULL,
        is_like    BOOLEAN NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id)    REFERENCES users(id),
        FOREIGN KEY (comment_id) REFERENCES comments(id)
    );`

	postLikes := `CREATE TABLE IF NOT EXISTS post_likes (
        id        INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id   INTEGER NOT NULL,
        post_id   INTEGER NOT NULL,
        is_like   BOOLEAN NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id),
        FOREIGN KEY (post_id) REFERENCES posts(id)
    );`

	postCategoryTable := `CREATE TABLE IF NOT EXISTS post_categories (
    post_id     INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);`

	notificationsTable := `CREATE TABLE IF NOT EXISTS notifications (
	        id            INTEGER PRIMARY KEY AUTOINCREMENT,
	        user_id       INTEGER NOT NULL,
	        initiator_id  INTEGER NOT NULL,
	        kind          TEXT    NOT NULL,      -- "like", "dislike", or "comment"
	        post_id       INTEGER NOT NULL,
	        comment_id    INTEGER,               -- nullable, only for comments
	        is_read       BOOLEAN DEFAULT 0,
	        created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	        FOREIGN KEY(user_id)      REFERENCES users(id),
	        FOREIGN KEY(initiator_id) REFERENCES users(id),
	        FOREIGN KEY(post_id)      REFERENCES posts(id),
	        FOREIGN KEY(comment_id)   REFERENCES comments(id)
	    );`

	for _, table := range []string{
		userTable, categoryTable, postTable,
		commentTable, sessionTable,
		commentLikes, postLikes, postCategoryTable,
		notificationsTable,
	} {
		if _, err := Database.Exec(table); err != nil {
			log.Printf("Error creating table: %v\n", err)
			return err
		}
	}

	return nil
}
