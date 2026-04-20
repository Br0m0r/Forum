package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/db"
)

func listenAndServe() {
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	registerRoutes()

	listenAndServe()
}
