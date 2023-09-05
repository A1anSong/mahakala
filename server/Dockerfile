FROM golang:alpine AS builder

WORKDIR /app
COPY . .

RUN go mod tidy && \
    go build -o server .

FROM alpine:latest

LABEL authors="A1an"

WORKDIR /app

COPY --from=builder /app/config.yaml .
COPY --from=builder /app/server .

EXPOSE 9999
ENTRYPOINT ./server