package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"strings"

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

	// ✅ Validate file size
	const maxFileSize = 10 * 1024 * 1024 // 10 MB
	if header.Size > maxFileSize {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// ✅ Validate file extension
	if len(header.Filename) < 4 || strings.ToLower(header.Filename[len(header.Filename)-4:]) != ".zip" {
		http.Error(w, "Only ZIP files allowed", http.StatusBadRequest)
		return
	}

	// Prepare upload directory
	os.MkdirAll("uploads", os.ModePerm)
	projectName := header.Filename[:len(header.Filename)-4]
	projectDir := filepath.Join("uploads", projectName)
	os.MkdirAll(projectDir, os.ModePerm)

	// Save ZIP
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

	log.Printf("[UPLOAD] Project: %s - Uploaded by: %s", projectName, r.RemoteAddr)

	// Unzip
	if err := builder.Unzip(zipPath, projectDir); err != nil {
		http.Error(w, "Unzip failed: "+err.Error(), http.StatusInternalServerError)
		log.Printf("[UNZIP] Project: %s - Failed: %v", projectName, err)
		return
	}
	log.Printf("[UNZIP] Project: %s - Unzip successful", projectName)

	projectRoot := builder.FindGoProjectRoot(projectDir)
	var bldr builder.Builder = &builder.GoBuilder{}

	// Build
	log.Printf("[BUILD] Project: %s - Starting build", projectName)
	if err := bldr.Build(projectRoot); err != nil {
		http.Error(w, "Build failed:\n"+err.Error(), http.StatusInternalServerError)
		log.Printf("[BUILD] Project: %s - Failed: %v", projectName, err)
		return
	}
	log.Printf("[BUILD] Project: %s - Build successful", projectName)

	// 🔹 Configurable timeout (default 5s)
	timeout := 5 * time.Second
	if tStr := r.URL.Query().Get("timeout"); tStr != "" {
		if tSec, err := strconv.Atoi(tStr); err == nil && tSec > 0 {
			timeout = time.Duration(tSec) * time.Second
		}
	}
	log.Printf("[RUN] Project: %s - Starting execution (timeout=%v)", projectName, timeout)

	// Run
	output, err := bldr.RunWithTimeout(projectRoot, timeout)
	if err != nil {
		http.Error(w, "Run failed:\n"+err.Error()+"\n\nOutput:\n"+output, http.StatusInternalServerError)
		log.Printf("[RUN] Project: %s - Failed: %v\nOutput:\n%s", projectName, err, output)
		return
	}

	log.Printf("[RUN] Project: %s - Execution finished successfully\nOutput:\n%s", projectName, output)
	w.Write([]byte("Execution Output:\n" + output))
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Host Go Land Server Running"))
	})

	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}