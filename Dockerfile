# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Guarantee both asset dirs exist regardless of git tracking
RUN mkdir -p templates static

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-w -s" -o wrenchlog .

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

RUN mkdir -p templates static /data

COPY --from=builder /app/wrenchlog .
COPY --from=builder /app/templates/ templates/
COPY --from=builder /app/static/    static/

EXPOSE 8080

ENV DB_PATH=/data/wrenchlog.db

CMD ["./wrenchlog"]
