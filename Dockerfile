FROM golang:1.26-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct \
    GOPRIVATE=github.com/zzznow \
    GONOSUMDB=*

WORKDIR /build/auth
COPY auth/ ./
COPY models/ ../models/
RUN go build -mod=mod -ldflags="-s -w" -o /app ./cmd

FROM alpine:3.23
WORKDIR /apps
ENV LANG=en_US.UTF-8

RUN echo "http://mirrors.tuna.tsinghua.edu.cn/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirrors.tuna.tsinghua.edu.cn/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    update-ca-certificates

COPY --from=builder /app .

EXPOSE 80
ENTRYPOINT ["./app"]
