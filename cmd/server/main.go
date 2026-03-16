package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sushilchlgn/hostgolang/internal/builder"
)
// UploadHandler handles ZIP uploads
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST allowed"))
		return
	}

	file, header, err := r.FormFile("project")
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), 400)
		return
	}
	defer file.Close()

	// Prepare upload folder
	os.MkdirAll("uploads", os.ModePerm)

	// Save the uploaded zip
	zipPath := filepath.Join("uploads", header.Filename)
	dst, err := os.Create(zipPath)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), 500)
		return
	}
	defer dst.Close()
	_, err = dst.ReadFrom(file)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), 500)
		return
	}

	// Extract zip to a dedicated folder
	projectName := header.Filename
	if len(projectName) > 4 && projectName[len(projectName)-4:] == ".zip" {
		projectName = projectName[:len(projectName)-4] // remove .zip
	}
	projectDir := filepath.Join("uploads", projectName)
	os.MkdirAll(projectDir, os.ModePerm)

	err = builder.Unzip(zipPath, projectDir)
	if err != nil {
		http.Error(w, "Unzip failed: "+err.Error(), 500)
		return
	}

	// Detect Go project root inside extracted folder
	projectRoot := builder.FindGoProjectRoot(projectDir)

	// Build the Go project
	goBuilder := builder.GoBuilder{}
	err = goBuilder.Build(projectRoot)
	if err != nil {
		http.Error(w, "Build failed: "+err.Error(), 500)
		return
	}

	// Run the Go binary
	err = goBuilder.Run(projectRoot)
	if err != nil {
		http.Error(w, "Run failed: "+err.Error(), 500)
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