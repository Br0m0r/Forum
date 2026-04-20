package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var Database *sql.DB

func InitDB() error {

	_, statErr := os.Stat("./db")
	if os.IsNotExist(statErr) {
		makeDirErr := os.MkdirAll("./db", 0755)
		if makeDirErr != nil {
			return fmt.Errorf("failed to create database directory: %v", makeDirErr)
		}
	}

	var openErr error
	Database, openErr = sql.Open("sqlite3", "./db/forum.db")
	if openErr != nil {
		return fmt.Errorf("failed to open database: %v", openErr)
	}

	pingErr := Database.Ping()
	if pingErr != nil {
		return fmt.Errorf("failed to connect to database: %v", pingErr)
	}

	_, foreignKeyErr := Database.Exec("PRAGMA foreign_keys = ON;")
	if foreignKeyErr != nil {
		return fmt.Errorf("failed to enable foreign keys: %v", foreignKeyErr)
	}

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
