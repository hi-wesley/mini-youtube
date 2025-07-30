FROM golang:1.23-bookworm AS builder
# Set the working directory to where the Go module root will be.
WORKDIR /src

# Copy the go.mod and go.sum files first to leverage Docker layer caching.
COPY go.mod go.sum ./

# Download dependencies.
RUN go mod download

# Copy the entire backend source code.
COPY . ./

# Build the application. The path is now relative to the module root.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /server ./cmd/server

# Start a new, smaller stage for the final image.
FROM debian:bookworm-slim

# Install ffmpeg and ca-certificates
RUN apt-get update && apt-get install -y ffmpeg ca-certificates --no-install-recommends && rm -rf /var/lib/apt/lists/*
ENV PORT=8080
EXPOSE 8080

# Copy the compiled application from the builder stage.
COPY --from=builder /server /server

# Copy the Firebase credentials file
COPY firebasekey.json /firebasekey.json

# Set the entrypoint for the container.
ENTRYPOINT ["/server"]