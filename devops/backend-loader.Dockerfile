# dockerfile
FROM golang:1.24-alpine
WORKDIR /app
COPY ./backend /app
RUN go build -o /app/shadowapi ./cmd/shadowapi
ENV SA_SKIP_WORKER=true

CMD ["/app/shadowapi", "loader"]