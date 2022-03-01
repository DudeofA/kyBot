FROM golang:alpine AS build

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod ./
COPY go.sum ./
COPY CHANGELOG.md /CHANGELOG.md

RUN go mod download

COPY *.go ./

RUN go build -o /kybot

##
## Deploy
##

FROM alpine:latest AS run

WORKDIR /

RUN apk add --no-cache tzdata

COPY --from=build /kybot /kybot
COPY --from=build /CHANGELOG.md /CHANGELOG.md
RUN mkdir -p /data

ENTRYPOINT ["/kybot"]
LABEL org.opencontainers.image.description "https://github.com/kylixor/kybot"