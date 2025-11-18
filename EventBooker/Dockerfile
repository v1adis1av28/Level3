FROM golang:1.25.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
RUN go mod tidy
COPY . .
RUN go build -o main ./cmd
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY internal/web/static ./internal/web/static
EXPOSE 8080
CMD ["./main"]
