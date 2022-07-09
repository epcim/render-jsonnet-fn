FROM golang:1.16-alpine3.13 as golang

ENV CGO_ENABLED=0
WORKDIR /go/src/

# DISABLED, not implemented
#COPY go.mod go.sum ./
#RUN go mod download
#COPY . .
#RUN go build -o /usr/local/bin/function ./

RUN go install -a github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@latest

#RUN apk update && apk add --no-cache curl

#############################################
FROM bitnami/jsonnet:latest as jsonnet


#############################################
FROM alpine:latest as render-jsonnet

#COPY --from=golang /usr/local/bin/function /usr/local/bin/function
COPY --from=golang  /go/bin/jb /usr/local/bin/jb
COPY --from=jsonnet /opt/bitnami/jsonnet/bin /usr/local/bin/jsonnet
COPY vendor /vendor

RUN apk update && apk add --no-cache git yq

COPY fnRenderJsonnet /usr/loca/bin/fnRenderJsonnet
COPY fnRenderJsonnet /usr/loca/bin/fnRenderJsonnet.sh
ENV PATH /usr/local/bin:$PATH

ARG DEBUG=False
ENV DEBUG $DEBUG

#ARG RENDER_TEMP=/tmp/render
#ENV RENDER_TEMP $RENDER_TEMP
#ARG SOURCES_DIR=/tmp/sources
#ENV SOURCES_DIR $SOURCES_DIR
#ARG UPDATE_SOURCE=False
#ENV UPDATE_SOURCE $UPDATE_SOURCE

ENTRYPOINT ["function"]
