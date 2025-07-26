# ---------- build stage ----------
FROM golang:1.22-bookworm AS builder
WORKDIR /src
RUN apt-get update && apt-get install -y ffmpeg

# Copy the Go modules files
COPY go.mod go.sum ./

# Download Go modules
COPY backend/. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /app ./cmd/server

# ---------- final image ----------
FROM gcr.io/distroless/base-debian12
ENV PORT=8080
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
