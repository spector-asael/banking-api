package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Counter holds our "Source of Truth"
// There is no database here, just a simple in-memory counter.
// In a real application, this would likely be stored in a database or cache like Redis.
// Using a Mutex ensures that if two students click at once, the count is safe.
// We use a Mutex to ensure that only one goroutine can access the counter at a time, preventing
// race conditions and ensuring that the count is accurate even when multiple requests
// are made simultaneously.
type Counter struct {
	mu    sync.Mutex
	Value int
}

var globalCounter = Counter{
	Value: 0,
}

func main() {
	mux := http.NewServeMux()

	// Serve static files (HTML, CSS, JS) from the "static" folder
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	// API Route: Increment count
	mux.HandleFunc("POST /api/increment", incrementCount)

	fmt.Println("Server starting at http://localhost:9000")
	if err := http.ListenAndServe(":9000", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func incrementCount(w http.ResponseWriter, r *http.Request) {
	globalCounter.mu.Lock()
	globalCounter.Value++
	current := globalCounter.Value
	globalCounter.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"value": current})
}
