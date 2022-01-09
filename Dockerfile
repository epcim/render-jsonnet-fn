FROM golang:1.16-alpine3.13
ENV CGO_ENABLED=0
WORKDIR /go/src/

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/function ./
RUN go install -a github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@latest

#RUN apk update && apk add --no-cache curl

#############################################

FROM alpine:3.13
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
COPY --from=0 /usr/local/bin/jb /usr/local/bin/jb

RUN apk update && apk add --no-cache git

ENV PATH /usr/local/bin:$PATH

ARG DEBUG=False
ENV DEBUG $DEBUG
ARG RENDER_TEMP=/tmp/render
ENV RENDER_TEMP $RENDER_TEMP
ARG SOURCES_DIR=/tmp/sources
ENV SOURCES_DIR $SOURCES_DIR
ARG UPDATE_SOURCE=False
ENV UPDATE_SOURCE $UPDATE_SOURCE

ENTRYPOINT ["function"]
