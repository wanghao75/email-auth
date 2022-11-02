FROM golang:alpine as builder

MAINTAINER wanghao<shalldows@163.com>

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct
WORKDIR $GOPATH/src/email-auth

RUN apk add --no-cache ca-certificates \
    tzdata \
    bash \
    bash-doc \
    bash-completion \
    && rm -rf /var/cache/apk/* \
    && update-ca-certificates

# 将当前目录同步到docker工作目录下
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM scratch

MAINTAINER wanghao<shalldows@163.com>

ENV GIN_MOD=release

WORKDIR /DockerTest

COPY --from=builder /go/src/email-auth/main .

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /usr/share/ca-certificates /usr/share/ca-certificates
COPY --from=builder /usr/local/share/ca-certificates /usr/local/share/ca-certificates
ENV TZ=Asia/Shanghai

EXPOSE 8080

ENTRYPOINT ["./main"]
