FROM golang:1.21.4-alpine3.17 AS builder

WORKDIR /build

COPY ./ /build

RUN go build -o dist/spooky-server ./cmd

FROM alpine:3.17

WORKDIR /opt/spooky-server

COPY --from=builder /build/dist/spooky-server /opt/spooky-server/spooky-server

ENTRYPOINT [ "/opt/spooky-server/spooky-server" ]
