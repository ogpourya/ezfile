# ezfile

Minimal, secure file upload server in Go for fast, on-the-fly usage.

## Install

```bash
GOPROXY=direct go install github.com/ogpourya/ezfile@latest
```

## Run

```bash
ezfile
```
Server starts on `:8080`.

## Upload

**From file:**
```bash
curl -F "file=@image.png" http://localhost:8080/
```

**From pipe:**
```bash
ls -la | curl -F "file=@-;filename=list.txt" http://localhost:8080/
```

Files are saved to your home directory as `filename.ezfile.type`.

## Security Note

This server is designed for quick, on-the-fly file uploads. It has **no authentication** and **no upload limit**, meaning it has **low security**. Use with caution and only in trusted environments or behind appropriate access controls.