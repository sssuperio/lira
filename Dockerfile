FROM node:22-alpine AS web-builder
WORKDIR /src
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && corepack install && pnpm install --frozen-lockfile
COPY . .
RUN pnpm run build

FROM golang:1.26-alpine AS go-builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
COPY --from=web-builder /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(cat VERSION 2>/dev/null || echo 'docker')" -o /lira .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
RUN mkdir -p /data && chown -R app:app /app /data
COPY --from=go-builder /lira /usr/local/bin/lira
USER app
VOLUME ["/data"]
ENV PORT=8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -qO- "http://127.0.0.1:${PORT:-8080}/healthz" >/dev/null || exit 1
CMD ["sh", "-c", "exec lira --addr :${PORT:-8080} --data-dir /data --allow-origin '*'"]
