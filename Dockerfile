FROM golang:1.25-alpine AS build

RUN apk update && apk add --no-cache git


WORKDIR /app

# RUN git clone https://github.com/michaelolawoye/url-shortener.git .
COPY . .

RUN go build


FROM alpine
WORKDIR /app
COPY --from=build /app/main .
ENV REDIS_HOST="host.docker.internal"

CMD ["./main"]
