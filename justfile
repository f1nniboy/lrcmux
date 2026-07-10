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

lint: lint-backend lint-frontend

lint-backend:
    golangci-lint run ./...

lint-frontend:
    cd frontend && pnpm lint

fix: fmt fix-backend fix-frontend

fix-backend:
    go run golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest -fix ./...
    go fix ./...
    golangci-lint run --fix ./...

fix-frontend:
    cd frontend && pnpm fix

fmt:
    treefmt

test:
    go test ./...

deploy-api:
    fly deploy --local-only

deploy-frontend:
    cd frontend && pnpm deploy

deploy: deploy-api deploy-frontend

clean:
    rm -rf frontend/build frontend/.svelte-kit lrcmux
