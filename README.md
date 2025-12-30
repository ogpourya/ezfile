# ezfile

Minimal file upload server in Go.

## Install

```bash
GOPROXY=direct go install github.com/ogpourya/ezfile@latest
```

## Usage

```bash
ezfile [--host <address>] [--port <port>] [--urlencoded]
```

- **File upload:** `curl -F "file=@image.png" http://localhost:8080/`
- **Piped upload:** `ls -la | curl -F "file=@-;filename=list.txt" http://localhost:8080/`
- **URL encoded:** `curl http://localhost:8080/ -d file=$(cat output.txt)` (requires `--urlencoded`)

Files are saved to `~/` with timestamped names and automatic extension detection.

## Security

**No authentication. No limits.** Use only in trusted networks.
