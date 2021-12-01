FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /kybot

##
## Deploy
##

FROM alpine:latest

WORKDIR /

COPY --from=build /kybot /kybot

ENTRYPOINT ["/kybot"]