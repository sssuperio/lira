FROM node:22-alpine AS web-builder
WORKDIR /src
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY . .
RUN pnpm run build

FROM golang:1.26-alpine AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(cat VERSION 2>/dev/null || echo 'docker')" -o /lira .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
RUN mkdir -p /app/data && chown -R app:app /app
COPY --from=go-builder /lira /usr/local/bin/lira
USER app
VOLUME ["/app/data"]
ENV PORT=8080
CMD ["sh", "-c", "exec lira --addr :${PORT:-8080} --data-dir /app/data --allow-origin '*'"]
