FROM golang:1.26-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct \
    GONOSUMDB=*

WORKDIR /build/auth
COPY auth/ ./
COPY models/ ../models/
RUN go build -mod=vendor -ldflags="-s -w" -o /app ./cmd

FROM alpine:3.23
WORKDIR /apps
ENV LANG=en_US.UTF-8

RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    update-ca-certificates

COPY --from=builder /app .

EXPOSE 80
ENTRYPOINT ["./app", "prod"]
