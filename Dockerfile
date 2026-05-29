FROM golang:1.26.3-alpine as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct \
    GONOSUMDB=*

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /build/auth
COPY auth/ ./
COPY models/ ../models/
RUN go build -mod=vendor -ldflags="-s -w" -o /app ./cmd

FROM alpine
WORKDIR /apps
ENV LANG en_US.UTF-8

COPY --from=builder /etc/localtime /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app .

EXPOSE 80
ENTRYPOINT ["./app", "prod"]
