FROM golang:1.26-bookworm AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct \
    GONOSUMDB=*

RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources && \
    apt-get update && \
    apt-get install -y --no-install-recommends tzdata ca-certificates && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build/auth
COPY auth/ ./
COPY models/ ../models/
RUN go build -mod=vendor -ldflags="-s -w" -o /app ./cmd

FROM alpine
WORKDIR /apps
ENV LANG en_US.UTF-8

COPY --from=builder /etc/localtime /etc/localtime
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app .

EXPOSE 80
ENTRYPOINT ["./app", "prod"]
