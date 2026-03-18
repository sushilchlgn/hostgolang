package main

import (
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

	os.MkdirAll("uploads", os.ModePerm)

	projectName := header.Filename
	if len(projectName) > 4 && projectName[len(projectName)-4:] == ".zip" {
		projectName = projectName[:len(projectName)-4]
	}

	projectDir := filepath.Join("uploads", projectName)
	os.MkdirAll(projectDir, os.ModePerm)

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

	if err := builder.Unzip(zipPath, projectDir); err != nil {
		http.Error(w, "Unzip failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	projectRoot := builder.FindGoProjectRoot(projectDir)

	var bldr builder.Builder = &builder.GoBuilder{}

	if err := bldr.Build(projectRoot); err != nil {
		http.Error(w, "Build failed:\n"+err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := bldr.Run(projectRoot)
	if err != nil {
		http.Error(w, "Run failed:\n"+err.Error()+"\n\nOutput:\n"+output, http.StatusInternalServerError)
		return
	}

	// ✅ Single response only
	w.Write([]byte("Execution Output:\n" + output))
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Host Go Land Server Running"))
	})

	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}