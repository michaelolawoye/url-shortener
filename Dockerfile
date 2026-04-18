FROM golang:1.25-alpine AS build

RUN apk update && apk add --no-cache git


WORKDIR /app

ENV REDIS_ADDR="host.docker.internal:6379"

RUN git clone https://github.com/michaelolawoye/url-shortener.git .

RUN go build

CMD ["./main"]
