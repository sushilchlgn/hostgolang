package main

import (
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Host Go Land Server Running"))
	})

	log.Println("Server running on port 8080")

	http.ListenAndServe(":8080", nil)
}