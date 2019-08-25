FROM golang:1.12-alpine as builder

ENV GO111MODULE=on GOPROXY=https://goproxy.io

RUN apk add --update --no-cache git

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY main main

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o build/main main/*.go

FROM alpine:3.10

ENV DNS_ENV=PRODUCTION \
    TYPE=A \
    INTERVAL=30 \
    TTL=600

RUN apk add --update --no-cache ca-certificates

COPY --from=builder /app/build/main /app/

ENTRYPOINT ["/app/main"]