package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	// Initial Set up: web server

	pingHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "pong!\n")
	}

	http.HandleFunc("/ping", pingHandler)
	log.Println("Listing for requests at http://localhost:8000/ping")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
