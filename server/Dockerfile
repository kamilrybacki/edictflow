# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files for both server and pkg
COPY server/go.mod server/go.sum ./server/
COPY pkg/go.mod ./pkg/

# Download dependencies
WORKDIR /app/server
RUN go mod download

# Copy source
WORKDIR /app
COPY server/ ./server/
COPY pkg/ ./pkg/

# Build master
WORKDIR /app/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /master ./cmd/master

# Build worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# Build legacy server (for backward compatibility)
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Master image
FROM alpine:3.23 AS master
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /master .
COPY --from=builder /go/bin/migrate .
COPY server/migrations ./migrations
COPY server/scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./master"]

# Worker image
FROM alpine:3.23 AS worker
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /worker .
EXPOSE 8081
CMD ["./worker"]

# Default image (legacy - combined server)
FROM alpine:3.23 AS server
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /server .
COPY --from=builder /go/bin/migrate .
COPY server/migrations ./migrations
COPY server/scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./server"]
