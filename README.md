# lrcmux

A lyrics aggregator. Fans out requests across multiple providers in parallel, picks the best result, and caches everything.

A public instance runs at **[lrcmux.dev](https://lrcmux.dev)**. The API docs are browsable at [lrcmux.dev/docs](https://lrcmux.dev/docs).

## Self-hosting

### Docker Compose

The easiest way to run the full stack (API + frontend + Redis):

```sh
git clone https://github.com/f1nniboy/lrcmux
cd lrcmux
cp config.example.toml config.toml
docker compose up
```

The API will be available at `http://localhost:8080` and the frontend at `http://localhost:3000`.

### Binary

```sh
git clone https://github.com/f1nniboy/lrcmux
cd lrcmux
cp config.example.toml config.toml
go run ./cmd/lrcmux -config config.toml
```

### Fly.io

The API and frontend are deployed as separate apps. Create both apps first:

```sh
fly launch --no-deploy --config fly/api.toml
fly launch --no-deploy --config fly/frontend.toml
fly secrets set REDIS_URL=redis://... --config fly/api.toml
```

Then deploy:

```sh
just deploy
```

## Configuration

See `config.example.toml` for all available options.
