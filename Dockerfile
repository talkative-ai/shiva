FROM golang:alpine

ADD . /go/src

RUN go install github.com/artificial-universe-maker/shiva

ENTRYPOINT /go/bin/shiva

EXPOSE 8080