# Stage 1: Build binary
FROM golang:alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o vvinamp-server main.go

# Stage 2: Runtime container
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (ffmpeg, yt-dlp, ca-certificates, tzdata)
RUN apk add --no-cache ffmpeg yt-dlp ca-certificates tzdata

# Set default media storage paths
RUN mkdir -p /app/media/downloads /app/media/hls

# Copy compiled binary from builder
COPY --from=builder /app/vvinamp-server /app/vvinamp-server

# Environment defaults
ENV PORT=8080 \
    GIN_MODE=release \
    FFMPEG_PATH=/usr/bin/ffmpeg \
    YT_DLP_PATH=/usr/bin/yt-dlp \
    STORAGE_PATH=/app/media/downloads

EXPOSE 8080

CMD ["/app/vvinamp-server"]
