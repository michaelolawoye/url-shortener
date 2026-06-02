FROM golang:1.25-alpine AS build

RUN apk update && apk add --no-cache git


WORKDIR /app

RUN git clone https://github.com/michaelolawoye/url-shortener.git .


# COPY . .

RUN go build ./cmd/backend


FROM alpine
WORKDIR /app
COPY --from=build /app/backend .
ENV REDIS_HOST="url-shortener-redis-service"

CMD ["ls"]
