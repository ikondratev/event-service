# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/http \
    ./cmd/http

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/worker \
    ./cmd/worker    

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/bin/http /app/http
COPY --from=builder /app/bin/worker /app/worker

USER appuser
EXPOSE 3333

CMD ["/app/http"]