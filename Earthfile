VERSION 0.8

FROM docker.io/library/golang:1.18-alpine

WORKDIR /repo

deps:
    COPY . .
    RUN mkdir -p build
    RUN go mod download
    
build:
    FROM +deps
    RUN env GOOS=linux GOARCH=amd64 go build -o build/

save:
    FROM +build
    SAVE ARTIFACT build/* AS LOCAL artifact/

all:
    BUILD +save