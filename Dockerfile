FROM golang:1.20

WORKDIR /app

COPY ./ .

RUN go build -o build/spooky-server ./cmd

ENTRYPOINT [ "/app/build/spooky-server" ]