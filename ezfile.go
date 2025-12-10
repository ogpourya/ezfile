package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"strings"
	"time"

	"flag"
	"mime"
)

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Define command-line flags for host and port
	host := flag.String("host", "", "Host address to listen on (default: all interfaces)")
	port := flag.String("port", "8080", "Port to listen on")
	urlEncoded := flag.Bool("urlencoded", false, "Enable URL encoded mode (expects application/x-www-form-urlencoded)")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uploadHandler(w, r, *urlEncoded)
	})

	listenAddr := fmt.Sprintf("%s:%s", *host, *port)

	log.Println("Starting ezfile server on:")
	if *host == "" || *host == "0.0.0.0" {
		// Print localhost
		log.Printf("- http://localhost:%s", *port)

		// Print other IPs
		ifaces, err := net.Interfaces()
		if err == nil {
			for _, i := range ifaces {
				addrs, err := i.Addrs()
				if err == nil {
					for _, addr := range addrs {
						var ip net.IP
						switch v := addr.(type) {
						case *net.IPNet:
							ip = v.IP
						case *net.IPAddr:
							ip = v.IP
						}
						// Print IPv4 and non-loopback (loopback covered by localhost)
						if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
							log.Printf("- http://%s:%s", ip.String(), *port)
						}
					}
				}
			}
		}

		// Fetch and print Public IP
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("https://ipv4.icanhazip.com/")
		if err == nil {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				publicIP := strings.TrimSpace(string(body))
				if publicIP != "" {
					log.Printf("- http://%s:%s (Public)", publicIP, *port)
				}
			}
		} else {
			// Fail silently or log debug info if needed, but user just wants it shown if available
		}
	} else {
		log.Printf("- http://%s:%s", *host, *port)
	}

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request, urlEncoded bool) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var src io.ReadSeeker
	var filename string

	if urlEncoded {
		// Parse form for URL encoded data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		content := r.FormValue("file")
		if content == "" {
			http.Error(w, "Invalid content. Use form field 'file'", http.StatusBadRequest)
			return
		}
		src = strings.NewReader(content)
		// No filename in urlencoded mode, so leave empty to trigger generation
		filename = ""
	} else {
		// Parse multipart form (no size limit enforced here, handled by io.Copy below)
		// 2. Retrieve the file
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Invalid file. Use form field 'file'", http.StatusBadRequest)
			return
		}
		defer file.Close()
		src = file
		filename = header.Filename
	}

	// 3. Security: Sanitize filename and detect extension if needed
	// filepath.Base prevents directory traversal
	if filename == "" || filename == "." || filename == "/" {
		// Detect extension
		head := make([]byte, 512)
		n, _ := src.Read(head)
		src.Seek(0, io.SeekStart) // Reset reader

		contentType := http.DetectContentType(head[:n])
		exts, _ := mime.ExtensionsByType(contentType)
		extension := ""
		if len(exts) > 0 {
			extension = exts[0]
			for _, ext := range exts {
				if ext == ".txt" {
					extension = ext
					break
				}
			}
		}
		filename = fmt.Sprintf("upload_%s%s", time.Now().Format("15-04-05"), extension)
	} else {
		filename = filepath.Base(filename)
	}

	// Additional sanitization: replace spaces and weird chars with underscores
	safeFilename := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, filename)

	// 4. Determine destination
	// Use current user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp if user lookup fails
		homeDir = "/tmp"
	}

	// Base path: <home>/<filename>
	finalPath := filepath.Join(homeDir, safeFilename)

	// 5. Save the file
	dst, err := os.Create(finalPath)
	if err != nil {
		http.Error(w, "Server error creating file", http.StatusInternalServerError)
		log.Printf("Error creating file %s: %v", finalPath, err)
		return
	}
	
	_, err = io.Copy(dst, src)
	dst.Close()
	if err != nil {
		http.Error(w, "Server error saving file", http.StatusInternalServerError)
		log.Printf("Error writing to file %s: %v", finalPath, err)
		return
	}

	log.Printf("Successfully saved: %s", finalPath)
	fmt.Fprintf(w, "Saved to: %s\n", filepath.Base(finalPath))
}
