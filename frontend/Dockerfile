# ---- build stage ----
FROM node:20-bookworm AS builder
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ---- serve stage ----
FROM gcr.io/distroless/nodejs20-debian12
ENV PORT=8080
WORKDIR /app
COPY --from=builder /app/dist /app/dist
COPY --from=builder /app/node_modules/.bin/serve /usr/local/bin/serve
CMD ["serve", "-s", "dist", "-l", "8080"]
