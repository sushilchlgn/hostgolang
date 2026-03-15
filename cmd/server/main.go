package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sushilchlgn/hostgolang/internal/builder"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST allowed"))
		return
	}

	// Parse file from form
	file, header, err := r.FormFile("project")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read file"))
		return
	}
	defer file.Close()

	// Save file to uploads folder
	os.MkdirAll("uploads", os.ModePerm)
	projectPath := filepath.Join("uploads", header.Filename)
	dst, err := os.Create(projectPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to save file"))
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	// --------------------------
	// Build and Run the project
	// --------------------------
	goBuilder := builder.GoBuilder{} // use package name 'builder'
	err = goBuilder.Build("uploads/" + header.Filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Build failed: " + err.Error()))
		return
	}

	err = goBuilder.Run("uploads/" + header.Filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Run failed: " + err.Error()))
		return
	}

	// Respond success
	fmt.Fprintf(w, "Uploaded, built, and running successfully: %s", header.Filename)
}
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Host Go Land Server Running"))
	})

	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}
