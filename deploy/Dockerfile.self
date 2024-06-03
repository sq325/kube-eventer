FROM golang:1.22.1 AS build-env
ADD . /src/github.com/sq325/kube-eventer
ENV GOPATH /:/src/github.com/sq325/kube-eventer/vendor
ENV GO111MODULE on
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src/github.com/sq325/kube-eventer
RUN apt-get update -y && apt-get install gcc ca-certificates
RUN make


FROM --platform=linux/amd64 alpine:latest

RUN apk --no-cache --update upgrade

COPY --from=build-env /src/github.com/sq325/kube-eventer/kube-eventer /
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV TZ "Asia/Shanghai"
RUN apk add --no-cache tzdata
#COPY deploy/entrypoint.sh /
RUN addgroup -g 1000 nonroot && \
    adduser -u 1000 -D -H -G nonroot nonroot && \
    chown -R nonroot:nonroot /kube-eventer
USER nonroot:nonroot

ENTRYPOINT ["/kube-eventer"]

