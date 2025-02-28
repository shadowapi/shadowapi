FROM golang:1.24-alpine

WORKDIR /app

# install bash in alpine without update
RUN apk add --no-cache bash

RUN \
  go install github.com/air-verse/air@v1.52.3 && \
  go install github.com/go-delve/delve/cmd/dlv@v1.22.1

CMD ["air", "-c", ".air.toml"]
