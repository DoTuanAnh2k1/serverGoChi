// Standalone frontend server for MGT Service.
// Serves the admin panel as a separate process, connecting to a remote mgt-service API.
//
// Usage:
//
//	MGT_API_URL=http://10.10.1.100:3000/aa go run ./cmd/frontend
//
// Environment variables:
//
//	MGT_API_URL    - URL of mgt-service API (required, e.g. http://host:3000/aa)
//	FRONTEND_PORT  - Port to listen on (default: 8080)
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	apiURL := os.Getenv("MGT_API_URL")
	if apiURL == "" {
		fmt.Println("MGT_API_URL is required")
		fmt.Println("Usage: MGT_API_URL=http://host:3000/aa go run ./cmd/frontend")
		os.Exit(1)
	}
	apiURL = strings.TrimRight(apiURL, "/")

	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "8080"
	}

	html, err := os.ReadFile("web/index.html")
	if err != nil {
		log.Fatalf("read web/index.html: %v", err)
	}

	page := strings.Replace(string(html), "{{API_BASE}}", apiURL, 1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
	})

	log.Printf("Frontend listening on :%s (API: %s)", port, apiURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
