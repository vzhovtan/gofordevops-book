package httpserver

import (
	"fmt"
	"log"
	"net/http"
)

func StartHttpServer() {
	// Define routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)

	// Start server
	port := ":9000"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// Handler for home page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Welcome to the Home Page</h1>")
}

// Handler for about page
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>About Page</h1>")
	fmt.Fprintf(w, "<p>This is a simple HTTP server built with Go</p>")
}
