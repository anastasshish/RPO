FROM golang:1.22-alpine AS builder
WORKDIR /src
// берет образ go
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=off

RUN printf '%s\n' \
    'https://dl-cdn.alpinelinux.org/alpine/v3.21/main' \
    'https://dl-cdn.alpinelinux.org/alpine/v3.21/community' > /etc/apk/repositories

RUN sh -c 'for i in 1 2 3 4 5; do apk add --no-cache git gcc musl-dev && exit 0; sleep 3; done; exit 1'

COPY go.mod ./
RUN sh -c 'for i in 1 2 3 4 5; do go mod download all && exit 0; sleep 5; done; exit 1'

COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
COPY docs ./docs

RUN sh -c 'for i in 1 2 3 4 5; do CGO_ENABLED=1 GOOS=linux go build -mod=mod -o /out/server ./cmd/server && exit 0; sleep 5; done; exit 1'

//Запускает приложение в другом образе
FROM alpine:3.20
WORKDIR /app

RUN printf '%s\n' \
    'https://dl-cdn.alpinelinux.org/alpine/v3.20/main' \
    'https://dl-cdn.alpinelinux.org/alpine/v3.20/community' > /etc/apk/repositories

RUN sh -c 'for i in 1 2 3 4 5; do apk add --no-cache nginx sqlite openssl ca-certificates && exit 0; sleep 3; done; exit 1'

COPY --from=builder /out/server /app/server
COPY migrations /app/migrations
COPY docs /app/docs
COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY docker/start.sh /start.sh

RUN sed -i 's/\r$//' /start.sh && \
    chmod +x /start.sh && \
    mkdir -p /etc/nginx/certs /app/data && \
    openssl req -x509 -nodes -days 365 \
    -newkey rsa:2048 \
    -keyout /etc/nginx/certs/server.key \
    -out /etc/nginx/certs/server.crt \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=Lab/CN=localhost"

ENV APP_ADDR=:8080
ENV DB_PATH=/app/data/app.db
ENV MIGRATIONS_PATH=/app/migrations
ENV JWT_SECRET=super-secret-change-me

EXPOSE 8888

ENTRYPOINT ["/start.sh"]
