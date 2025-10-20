FROM golang:1.25.1-alpine AS builder

ENV TZ=Europe/Moscow
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o /app/service ./cmd

FROM alpine:latest
WORKDIR /app

# Простая установка tzdata - этого обычно достаточно
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow

COPY --from=builder /app/service .
COPY --from=builder /app/config ./config

EXPOSE 8080
CMD ["./service"]