package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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

	var displayIP string = "localhost"
	log.Println("Starting ezfile server on:")
	if *host == "" || *host == "0.0.0.0" {
		log.Printf("- http://localhost:%s (Private)", *port)

		// 1. Get Local IPs and mark them
		localIPs := make(map[string]bool)
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
						if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
							ipStr := ip.String()
							localIPs[ipStr] = true
							label := ""
							if ip.IsPrivate() {
								label = " (Private)"
								if displayIP == "localhost" {
									displayIP = ipStr
								}
							} else {
								label = " (Public)"
								displayIP = ipStr
							}
							log.Printf("- http://%s:%s%s", ipStr, *port, label)
						}
					}
				}
			}
		}

		// 2. Fetch external Public IP (ignoring proxy)
		client := http.Client{
			Timeout: 2 * time.Second,
			Transport: &http.Transport{
				Proxy: nil, // This explicitly disables proxy for this request
			},
		}
		resp, err := client.Get("https://ipv4.icanhazip.com/")
		if err == nil {
			defer resp.Body.Close()
			if body, err := io.ReadAll(resp.Body); err == nil {
				publicIP := strings.TrimSpace(string(body))
				if publicIP != "" {
					// Only print if not already listed by local interfaces
					if !localIPs[publicIP] {
						log.Printf("- http://%s:%s (Public)", publicIP, *port)
					}
					displayIP = publicIP
				}
			}
		}
	} else {
		log.Printf("- http://%s:%s", *host, *port)
		displayIP = *host
	}

	fmt.Println("\nUsage examples:")
	fmt.Printf("  curl -F \"file=@image.png\" http://%s:%s/\n", displayIP, *port)
	fmt.Printf("  ls -la | curl -F \"file=@-;filename=list.txt\" http://%s:%s/\n", displayIP, *port)
	if *urlEncoded {
		fmt.Printf("  curl http://%s:%s/ -d file=$(cat /tmp/output)\n", displayIP, *port)
	}
	fmt.Println()

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
		filename = ""
	} else {
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Invalid file. Use form field 'file'", http.StatusBadRequest)
			return
		}
		defer file.Close()
		src = file
		filename = header.Filename
	}

	if filename == "" || filename == "." || filename == "/" {
		head := make([]byte, 512)
		n, _ := src.Read(head)
		src.Seek(0, io.SeekStart)

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

	safeFilename := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, filename)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}

	finalPath := filepath.Join(homeDir, safeFilename)

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
