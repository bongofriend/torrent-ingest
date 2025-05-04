FROM golang:1.24-alpine AS builder
USER root
WORKDIR /home/builder

COPY ./server /home/builder/torrent-ingest
WORKDIR /home/builder/torrent-ingest
RUN set -Eeux && \
    go mod download && \
    go mod verify

RUN go build -o app

FROM alpine:3.21
USER root
WORKDIR /home/app
RUN apk --no-cache add curl
COPY --from=builder /home/builder/torrent-ingest/app .
ENTRYPOINT [ "./app" ]