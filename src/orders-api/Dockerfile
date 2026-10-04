FROM golang:1.27-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

COPY . .
# Build the binary with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o orders-api main.go

# Minimal runtime; application dependencies still require vulnerability scans.
FROM scratch

WORKDIR /
COPY --from=builder /app/orders-api /orders-api
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Run as non-root user (nobody:nobody)
USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/orders-api"]
