# Dockerfile
FROM golang:1.25.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/order-service .
EXPOSE 8081
CMD ["./order-service"]
