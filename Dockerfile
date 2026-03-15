# Multi-stage build for WrenchLog (SQLite requires CGO)
FROM golang:1.20-alpine AS build

RUN apk add --no-cache build-base sqlite-dev git
WORKDIR /src

# dependency download
COPY go.mod go.sum ./
RUN go mod download

# copy source and build
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/wrenchlog

# Runtime image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates sqlite

COPY --from=build /app/wrenchlog /app/wrenchlog
WORKDIR /app

# Include templates/static so the binary can find them at /app/templates
COPY templates ./templates
COPY static ./static

EXPOSE 8080
CMD ["/app/wrenchlog"]
