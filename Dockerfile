FROM golang:alpine AS build

WORKDIR /app

RUN apk add gcc musl-dev

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./
COPY commands/*.go ./commands/
COPY config/*.go ./config/
COPY handlers/*.go ./handlers/
COPY kyDB/*.go ./kyDB/
COPY servers/*.go ./servers/

RUN go build -o /kybot

##
## Deploy
##

FROM alpine:latest

WORKDIR /

RUN apk update && apk add --no-cache tzdata

COPY --from=build /kybot /kybot
RUN mkdir -p /data

ENTRYPOINT ["/kybot"]