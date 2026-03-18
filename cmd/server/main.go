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

	// Prepare uploads folder
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

	// Unzip the project
	if err := builder.Unzip(zipPath, projectDir); err != nil {
		http.Error(w, "Unzip failed: "+err.Error(), http.StatusInternalServerError)
		log.Printf("[UNZIP] Project: %s - Failed: %v", projectName, err)
		return
	}
	log.Printf("[UNZIP] Project: %s - Unzip successful", projectName)

	// Detect Go project root
	projectRoot := builder.FindGoProjectRoot(projectDir)

	var bldr builder.Builder = &builder.GoBuilder{}

	// 🔹 Build
	log.Printf("[BUILD] Project: %s - Starting build", projectName)
	if err := bldr.Build(projectRoot); err != nil {
		http.Error(w, "Build failed:\n"+err.Error(), http.StatusInternalServerError)
		log.Printf("[BUILD] Project: %s - Failed: %v", projectName, err)
		return
	}
	log.Printf("[BUILD] Project: %s - Build successful", projectName)

	// 🔹 Run
	log.Printf("[RUN] Project: %s - Starting execution", projectName)
	output, err := bldr.Run(projectRoot)
	if err != nil {
		http.Error(w, "Run failed:\n"+err.Error()+"\n\nOutput:\n"+output, http.StatusInternalServerError)
		if output == "" {
			log.Printf("[RUN] Project: %s - Failed: %v", projectName, err)
		} else {
			log.Printf("[RUN] Project: %s - Failed: %v\nOutput:\n%s", projectName, err, output)
		}
		return
	}

	log.Printf("[RUN] Project: %s - Execution finished successfully\nOutput:\n%s", projectName, output)
	w.Write([]byte("Execution Output:\n" + output))
}

func main() {
	// Add timestamp and microseconds for precise logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Host Go Land Server Running"))
	})

	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}