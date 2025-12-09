# Build stage
FROM golang:alpine AS builder
WORKDIR /app

# Copy module files
COPY go.mod ./
# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 ensures a static binary, easier for Alpine
RUN CGO_ENABLED=0 go build -o ezfile ezfile.go

# Final stage
FROM alpine:latest

# Install 'file' utility (libmagic) required for file type detection
RUN apk add --no-cache file

# Create a non-root user for security
RUN adduser -D ezuser
USER ezuser
WORKDIR /home/ezuser
VOLUME /uploads

# Copy binary from builder
COPY --from=builder /app/ezfile /usr/local/bin/ezfile

# Expose port
EXPOSE 8080

# Run the server
CMD ["ezfile"]
