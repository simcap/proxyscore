FROM golang

RUN mkdir -p /go/src/github.com/simcap/proxyscore
ADD . /go/src/github.com/simcap/proxyscore

RUN go install github.com/simcap/proxyscore

ENTRYPOINT /go/bin/proxyscore

EXPOSE 4673
