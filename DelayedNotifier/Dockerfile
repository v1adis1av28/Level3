FROM golang:1.25.1-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o /app/service ./cmd

FROM alpine:latest
WORKDIR /app


COPY --from=builder /app/service .

EXPOSE 8080
CMD ["./service"]