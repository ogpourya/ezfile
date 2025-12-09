# ezfile - Simple, Secure File Upload Server

ezfile is a tiny, secure, and optimized Go-based file upload server designed to run efficiently within Docker containers. It provides a straightforward way to accept file uploads via `curl` POST requests, ensuring security and proper file handling.

## Features

*   **Minimalistic:** Single Go file, few dependencies.
*   **Secure:**
    *   File size limits (50MB).
    *   Sanitized filenames to prevent directory traversal.
    *   Replaces problematic characters with underscores.
*   **Optimized:** Determines file type using the `file` command.
*   **Container-Ready:** Includes a `Dockerfile` for easy deployment.
*   **Logging:** Logs requests safely with timestamps.

## Installation

To install the `ezfile` server locally, you can use `go install`:

```bash
GOPROXY=direct go install github.com/ogpourya/ezfile@latest
```

This will install the `ezfile` executable in your `$GOPATH/bin` directory.

## Usage

### Running the Server Locally

To run the server directly from the source:

```bash
go run ezfile.go
```

The server will start on `http://localhost:8080`.

### Building and Running with Docker

1.  **Build the Docker image:**

    ```bash
    docker build -t ezfile .
    ```

2.  **Run the Docker container (with volume mount for persistent storage):**

    ```bash
    # Replace /path/to/your/host/uploads with your desired host directory
    docker run --rm -p 8080:8080 -v "$(pwd)/uploads_data":/uploads ezfile
    ```
    The `--rm` flag ensures the container is automatically removed when it exits.
    The `-v "$(pwd)/uploads_data":/uploads` flag mounts a local directory named `uploads_data` (relative to your current terminal location) into the container's `/uploads` directory, making uploaded files persistent on your host machine. You can change `$(pwd)/uploads_data` to any directory on your host.

### Uploading Files with `curl`

Files are saved to your user's home directory (or `/tmp` if home directory is not found in the container) with `.ezfile.` appended to their original name, followed by their detected MIME type.

#### Uploading from a Local File Path

```bash
# Create a dummy file for testing
echo "Hello, ezfile!" > hello.txt

# Upload the file
curl -F "file=@hello.txt" http://localhost:8080/
```

#### Uploading via Pipe

```bash
# Example: Uploading the output of `ls -la` as `listing.txt`
ls -la | curl -F "file=@-;filename=listing.txt" http://localhost:8080/
```

## Security Notes

*   The server accepts uploads without authentication. It is intended for environments where this is acceptable or where network access is controlled.
*   File size is limited to 50MB (`MaxUploadSize`).
*   Filenames are sanitized to prevent directory traversal (`filepath.Base`) and problematic characters are replaced with underscores.
*   Files are saved to the `/uploads` directory within the container. When running with Docker and a volume is mounted (as shown in the usage section), these files will persist on the host system in the mounted directory. If no volume is mounted, files are stored ephemerally within the container.

## Development

### Dependencies

This project aims for minimal dependencies. The primary external tool used is the `file` command-line utility for MIME type detection, which is installed in the Docker image.

### Logging

Requests are logged to standard output with timestamps and file information.

```
# Example log entry:
2025/12/09 17:54:47 ezfile.go:127: Successfully saved: /root/testfile.txt.ezfile.text-plain (Type: text-plain)
```