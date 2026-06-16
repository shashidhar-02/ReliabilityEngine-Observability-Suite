FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
# Download dependencies
RUN go mod download

COPY . .
# Build the binary with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o orders-api main.go

# Use distroless for secure runtime
FROM gcr.io/distroless/static-debian11

WORKDIR /
COPY --from=builder /app/orders-api /orders-api

# Run as non-root user (nobody:nobody)
USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/orders-api"]
