# Build stage
FROM golang:1.21-alpine as builder
WORKDIR /app
COPY main.go .
RUN go mod init tbd-demo && \
    go build -o payment-service main.go

# Runtime stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/payment-service /usr/local/bin/
EXPOSE 8080
CMD ["payment-service"]
