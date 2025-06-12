package main

import (
	"fmt"
	"log"
	"modularMidiGoApp/backend/httpHandler"
	"strings"

	//"modularMidiGoApp/backend/usbUtility"
	"net/http"
	"os"
	"path/filepath"
)

// Executes first and prepares:
// - Starts HTTP handler
func main() {
	startHttpHandler()
}

// Returns the root path
func FindRootPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exePath)
	parentDir := filepath.Dir(dir)
	return parentDir
}

// Starts the HTTP handler (/httpHandler/handler.go)
func startHttpHandler() {
	httpPort = 
	mux := http.NewServeMux()
	mux.HandleFunc("/user", handlers.UserHandler)

	log.Printf("Server running on :%d", httpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux)
	if err != nil {
		log.Fatal(err)
	}
}