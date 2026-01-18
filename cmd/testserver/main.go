package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("cmd/testserver/static"))
	mux.Handle("/", fs)

	// Login endpoint
	mux.HandleFunc("/login", loginHandler)

	port := ":8082"
	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("Test Server starting on http://localhost%s\n", port)
	log.Fatal(server.ListenAndServe())
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Simulate processing time
	time.Sleep(500 * time.Millisecond)

	if password == "secret123" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<div id="response" class="success-message">
			<i class="fas fa-check-circle"></i>
			Welcome back, %s! Authentication successful.
		</div>`, username)
	} else {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<div id="response" class="error-message">
			<i class="fas fa-exclamation-triangle"></i>
			Access Denied. Incorrect password.
		</div>`)
	}
}
