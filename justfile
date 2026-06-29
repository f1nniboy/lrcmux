default: build

build: build-frontend build-backend

build-frontend:
    cd frontend && pnpm install --frozen-lockfile && pnpm build

build-backend:
    go build -ldflags="-X 'github.com/f1nniboy/lrcmux/internal/meta.Commit=$(git rev-parse --short HEAD)'" -o lrcmux ./cmd/lrcmux

dev-frontend:
    cd frontend && pnpm dev

dev-backend:
    go run ./cmd/lrcmux --config config.toml

logo:
    mkdir -p frontend/static
    inkscape assets/logo.svg --export-text-to-path --export-type=svg --export-filename=frontend/static/logo.svg
    npx svgo frontend/static/logo.svg --precision 2

test:
    go test ./internal/...

clean:
    rm -rf frontend/build frontend/.svelte-kit lrcmux
