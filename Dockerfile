FROM golang:alpine

RUN apk add --update git

COPY docker.gitconfig /root/.gitconfig

RUN go get github.com/talkative-ai/shiva

ENTRYPOINT /go/bin/shiva

EXPOSE 8080