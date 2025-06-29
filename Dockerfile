# Dockerfile for your Go API
FROM golang:1.23.4 AS builder
# Set necessary environment variables
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./
COPY go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main ./cmd/api
# Use a minimal image to run the application
FROM alpine:latest

WORKDIR /root/

# Copy the compiled Go application from the builder stage
COPY --from=builder /app/main .

COPY openapi.yaml .
# Expose the port your Go API listens on (e.g., 8080)
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
