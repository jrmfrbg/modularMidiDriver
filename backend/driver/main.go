package main

import (
	"fmt"
	"log"
	handlers "modularMidiGoApp/backend/httpHandler"
	"net/http"
)

// Executes first and prepares:
// - Starts HTTP handler
func main() {
	startHttpHandler()
}

// Returns the root path

// Starts the HTTP handler (/httpHandler/handler.go)
func startHttpHandler() {
	httpPort := LoadHTTPconf()
	mux := http.NewServeMux()
	mux.HandleFunc("/user", handlers.UserHandler)

	log.Printf("Server running on :%d", httpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux)
	if err != nil {
		log.Fatal(err)
	}
}
