FROM node:22-alpine AS frontend
WORKDIR /app/frontend
RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml* frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/build ./frontend/build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o lrcmux ./cmd/lrcmux

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/lrcmux /lrcmux
COPY config.prod.toml /etc/lrcmux/config.toml
EXPOSE 8080
ENTRYPOINT ["/lrcmux", "-config", "/etc/lrcmux/config.toml"]
