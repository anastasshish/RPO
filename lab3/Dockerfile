FROM golang:1.22-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /src/backend
COPY backend/go.mod ./
RUN go mod download
COPY backend/ ./

RUN go mod tidy && CGO_ENABLED=1 GOOS=linux go build -o /out/server ./cmd/server

FROM node:22-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY frontend/ ./
ENV NODE_ENV=production
RUN npm run build

FROM nginx:1.27-alpine
RUN apk add --no-cache sqlite-libs
WORKDIR /app
COPY --from=backend-builder /out/server /app/server
COPY backend/migrations /app/migrations
COPY --from=frontend-builder /src/frontend/dist /usr/share/nginx/html
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
COPY certs /etc/nginx/certs
COPY scripts/entrypoint.sh /entrypoint.sh
ENV DB_PATH=/app/data/transport.db MIGRATIONS_DIR=/app/migrations

RUN mkdir -p /app/data && chmod +x /entrypoint.sh
EXPOSE 8888
CMD ["/entrypoint.sh"]
