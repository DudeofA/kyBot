FROM golang:alpine AS build

WORKDIR /app

RUN apk update && apk add gcc musl-dev

COPY go.mod ./
COPY go.sum ./
COPY CHANGELOG.md /CHANGELOG.md

RUN go mod download

COPY *.go ./
COPY commands/*.go ./commands/
COPY config/*.go ./config/
COPY handlers/*.go ./handlers/
COPY kyDB/*.go ./kyDB/
COPY component/*.go ./component/
COPY update/*.go ./update/

RUN go build -o /kybot

##
## Deploy
##

FROM alpine:latest AS run

WORKDIR /

RUN apk update && apk add --no-cache tzdata

COPY --from=build /kybot /kybot
COPY --from=build /CHANGELOG.md /CHANGELOG.md
RUN mkdir -p /data

ENTRYPOINT ["/kybot"]