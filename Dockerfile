FROM golang:1.26.3-alpine as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=off \
    GONOSUMDB=*

WORKDIR /build
COPY . .
RUN go build -mod=vendor -ldflags="-s -w" -o app ./auth/cmd

FROM alpine:3.21
RUN apk --no-cache add tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    update-ca-certificates

WORKDIR /apps
ENV LANG en_US.UTF-8

COPY --from=builder /build/app .
COPY --from=builder /etc/localtime /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 80
ENTRYPOINT ["./app", "prod"]
