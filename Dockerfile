FROM node:22-alpine AS frontend-build
WORKDIR /app
RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml* frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm svelte-kit sync && pnpm build

FROM golang:alpine AS api-build
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
      -X 'github.com/f1nniboy/lrcmux/internal/meta.Commit=$(git rev-parse --short HEAD)' \
      -X 'github.com/f1nniboy/lrcmux/internal/meta.Version=$(git describe --tags --always 2>/dev/null || echo dev)'" \
    -o lrcmux ./cmd/lrcmux

FROM gcr.io/distroless/static-debian12 AS api
COPY --from=api-build /app/lrcmux /lrcmux
EXPOSE 8080
ENTRYPOINT ["/lrcmux", "-config", "/etc/lrcmux/config.toml"]

FROM node:22-alpine AS frontend
COPY --from=frontend-build /app/build /app
EXPOSE 3000
CMD ["node", "/app"]
