FROM golang:alpine AS builder

WORKDIR /app
COPY . .

RUN go mod tidy && \
    go build -o server .

FROM alpine:latest
LABEL authors="A1an"

RUN apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

WORKDIR /app

COPY --from=builder /app/config.yaml .
COPY --from=builder /app/server .

EXPOSE 9999
ENTRYPOINT ./server