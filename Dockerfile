FROM golang:1.17 as builder

MAINTAINER wanghao<shalldows@163.com>

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct
WORKDIR $GOPATH/src/email-auth

# RUN apk add --no-cache ca-certificates \
#     tzdata \
#     bash \
#     bash-doc \
#     bash-completion \
#     && rm -rf /var/cache/apk/* \
#     && update-ca-certificates

# 将当前目录同步到docker工作目录下
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM ubuntu:20.04

MAINTAINER wanghao<shalldows@163.com>

ENV GIN_MOD=release

WORKDIR /DockerTest

COPY --from=builder /go/src/email-auth/main .

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

RUN apt install -y curl

ENV TZ=Asia/Shanghai

EXPOSE 8080

ENTRYPOINT ["./main"]
