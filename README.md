# ezfile

Minimal file upload server in Go for fast, on-the-fly usage.

## Install

```bash
GOPROXY=direct go install github.com/ogpourya/ezfile@latest
```

## Run

```bash
ezfile [--host <address>] [--port <port>]
```

Server starts on `:8080` by default. You can specify a host address and/or port:

*   `ezfile --port 9000`
*   `ezfile --host 127.0.0.1 --port 8081`

## Upload

**From file:**
```bash
curl -F "file=@image.png" http://localhost:8080/
```

**From pipe:**
```bash
ls -la | curl -F "file=@-;filename=list.txt" http://localhost:8080/
```

Files are saved to your home directory as `filename`.

## Security Note

This server is designed for quick, on-the-fly file uploads. It has **no authentication** and **no upload limit**, meaning it has **low security**. Use with caution and only in trusted environments or behind appropriate access controls.