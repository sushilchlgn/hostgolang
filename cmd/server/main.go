package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sushilchlgn/hostgolang/internal/builder"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("project")
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Prepare uploads folder
	os.MkdirAll("uploads", os.ModePerm)

	// Prepare project folder
	projectName := header.Filename
	if len(projectName) > 4 && projectName[len(projectName)-4:] == ".zip" {
		projectName = projectName[:len(projectName)-4]
	}
	projectDir := filepath.Join("uploads", projectName)
	os.MkdirAll(projectDir, os.ModePerm)

	// Save uploaded ZIP
	zipPath := filepath.Join(projectDir, header.Filename)
	dst, err := os.Create(zipPath)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract ZIP into project folder
	if err := builder.Unzip(zipPath, projectDir); err != nil {
		http.Error(w, "Unzip failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Detect Go project root
	projectRoot := builder.FindGoProjectRoot(projectDir)

	// Use Builder interface
	var bldr builder.Builder = &builder.GoBuilder{}

	// Build project
	if err := bldr.Build(projectRoot); err != nil {
		http.Error(w, "Build failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Run project
	if err := bldr.Run(projectRoot); err != nil {
		http.Error(w, "Run failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Project deployed successfully")
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Host Go Land Server Running"))
	})

	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}