# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# -o /app/ollama-api-proxy: specifies the output file name.
# -ldflags="-w -s": reduces the size of the binary by removing debug information.
# CGO_ENABLED=0: disables CGO to create a static binary.
# Declare the build argument for the target architecture.
# Docker's buildx will automatically set this to the target architecture (e.g., amd64, arm64).
ARG TARGETARCH

# Build the application
# -o /app/ollama-api-proxy: specifies the output file name.
# -ldflags="-w -s": reduces the size of the binary by removing debug information.
# CGO_ENABLED=0: disables CGO to create a static binary.
# GOOS=linux GOARCH=${TARGETARCH}: specifies the target operating system and architecture.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
  go build -o /app/ollama-api-proxy \
  -ldflags="-w -s" \
  src/cmd/main/main.go

# Stage 2: Create the final image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/ollama-api-proxy .

# Copy the .env.example file. The user can mount a .env file to override it.
# COPY .env.example .env
COPY models.yml ./models.yml

ENV GIN_MODE=release

# Expose the port the app runs on
EXPOSE 11434

# Command to run the application
CMD ["./ollama-api-proxy"]