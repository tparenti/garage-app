# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

# gcc + musl-dev needed to compile go-sqlite3 (CGO)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Download dependencies first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source and assets, then build
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-w -s" -o wrenchlog .

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates for any outbound HTTPS; tzdata for correct timestamps
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary and assets all from the builder — avoids missing-dir errors
# from the build context when static/ is empty or not tracked by git
COPY --from=builder /app/wrenchlog .
COPY --from=builder /app/templates/ templates/
COPY --from=builder /app/static/    static/

# Data directory — mount a volume here to persist the SQLite database
RUN mkdir -p /data

EXPOSE 8080

# Pass DB path via environment variable (main.go reads DB_PATH, falls back to ./wrenchlog.db)
ENV DB_PATH=/data/wrenchlog.db

CMD ["./wrenchlog"]
