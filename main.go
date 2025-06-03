package main

import (
	"fmt"
	"forum/db"
	"log"
	"net/http"
)

func listenAndServe() {
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	// ListenAndServe takes two arguements, the port and an http.Handler
	// when we pass nil for the handler argument, we are telling Go to use the default ServeMux
	// the ServeMux is Go's built-in request multiplexer (router)
}

func main() {
	err := db.InitDB() // <== sqlite.go
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	registerRoutes() // <= routes.go

	listenAndServe()
}
