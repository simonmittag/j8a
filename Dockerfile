# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY . .

ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN go build -o j8a ./cmd/j8a/main.go

# Stage 2: Minimal runtime image
FROM alpine:latest

WORKDIR /

COPY --from=builder /app/j8a /j8a

ENV LOGLEVEL="DEBUG"

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/j8a"]