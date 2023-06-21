FROM golang:alpine3.18

RUN apk update && \
    apk add git

WORKDIR /go/src/mongo-log-driver

COPY . /go/src/mongo-log-driver/
RUN go mod tidy
RUN go get

RUN go build --ldflags '-extldflags "-static"' -o /usr/bin/mongo-log-driver