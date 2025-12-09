package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"strings"
	"time"
)

const (
	// MaxUploadSize limits the size of the upload to 50MB
	MaxUploadSize = 50 * 1024 * 1024
)

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/", uploadHandler)

	port := ":8080"
	log.Printf("Starting ezfile server on port %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Security: Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Error(w, "File too large or invalid format", http.StatusBadRequest)
		log.Printf("Upload rejected: %v", err)
		return
	}

	// 2. Retrieve the file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file. Use form field 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Security: Sanitize filename
	// filepath.Base prevents directory traversal
	filename := filepath.Base(header.Filename)
	if filename == "." || filename == "/" || filename == "" {
		filename = fmt.Sprintf("upload_%d", time.Now().Unix())
	}
	// Additional sanitization: replace spaces and weird chars with underscores
	safeFilename := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, filename)

	// 4. Determine destination
	// Save files to a dedicated /uploads directory inside the container.
	// This directory will be mounted as a volume by Docker.
	homeDir := "/uploads"
	// Ensure the upload directory exists
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		http.Error(w, "Server error preparing upload directory", http.StatusInternalServerError)
		log.Printf("Error creating upload directory %s: %v", homeDir, err)
		return
	}

	// Base path: <home>/<filename>.ezfile.
	// We will append the file type later
	tempPath := filepath.Join(homeDir, safeFilename+".ezfile")

	// 5. Save the file
	dst, err := os.Create(tempPath)
	if err != nil {
		http.Error(w, "Server error creating file", http.StatusInternalServerError)
		log.Printf("Error creating file %s: %v", tempPath, err)
		return
	}
	
	_, err = io.Copy(dst, file)
	dst.Close() // Close explicitly to flush before running 'file' command
	if err != nil {
		http.Error(w, "Server error saving file", http.StatusInternalServerError)
		log.Printf("Error writing to file %s: %v", tempPath, err)
		return
	}

	// 6. Identify file type
	// Run `file --brief --mime-type <file>`
	cmd := exec.Command("file", "--brief", "--mime-type", tempPath)
	out, err := cmd.Output()
	fileType := "unknown"
	if err == nil {
		fileType = strings.TrimSpace(string(out))
		// Sanitize mime type (e.g. "image/png" -> "image-png")
		fileType = strings.ReplaceAll(fileType, "/", "-")
	} else {
		log.Printf("Warning: 'file' command failed: %v", err)
	}

	// 7. Rename to include type
	// Final format: <original>.ezfile.<type>
	finalPath := tempPath + "." + fileType
	if err := os.Rename(tempPath, finalPath); err != nil {
		log.Printf("Error renaming file to %s: %v", finalPath, err)
		finalPath = tempPath // Fallback to tempPath
	}

	log.Printf("Successfully saved: %s (Type: %s)", finalPath, fileType)
	fmt.Fprintf(w, "Saved to: %s\n", filepath.Base(finalPath))
}
