# syntax=docker/dockerfile:1
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dumb-proxy-server ./cmd/server

FROM alpine:3.24

WORKDIR /root/

COPY --from=builder /app/dumb-proxy-server .

EXPOSE 8080

CMD ["./dumb-proxy-server"]
