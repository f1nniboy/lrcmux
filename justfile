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
    resvg frontend/static/logo.svg frontend/static/logo.png --width 512 --height 512
    magick frontend/static/logo.png -define icon:auto-resize=256,128,64,48,32,16 frontend/static/favicon.ico

test:
    go test ./internal/...

deploy-api:
    fly deploy --config fly/api.toml --local-only

deploy-frontend:
    fly deploy --config fly/frontend.toml --local-only

deploy: deploy-api deploy-frontend

clean:
    rm -rf frontend/build frontend/.svelte-kit lrcmux
