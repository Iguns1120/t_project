# Build Stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies (if any needed, e.g., for cgo)
# RUN apk add --no-cache build-base

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
# -o main: output binary name
# cmd/server/main.go: entry point
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

# Run Stage
FROM alpine:latest

WORKDIR /root/

# Copy the Pre-built binary from the builder stage
COPY --from=builder /app/main .

# Copy configuration file
# Assuming the app looks for config in ./configs/config.yaml relative to execution dir
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/docs ./docs

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
