FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN go build -o server .

FROM alpine:latest
LABEL authors="A1an"

RUN apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

WORKDIR /app

COPY --from=builder /app/server .

VOLUME ["/app/config.yaml","/app/log"]
EXPOSE 9999
ENTRYPOINT ./server