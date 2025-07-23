# ---------- build stage ----------
FROM golang:1.22-bookworm AS builder
WORKDIR /src
COPY backend/go.* ./
RUN go mod download
COPY backend/. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /app ./cmd/server

# ---------- final image ----------
FROM gcr.io/distroless/base-debian12
ENV PORT=8080
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
