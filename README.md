# lrcmux

A lyrics aggregator. Fans out requests across multiple providers in parallel, picks the best result, and caches everything.

A public instance runs at **[lrcmux.dev](https://lrcmux.dev)**. The API docs are browsable at [lrcmux.dev/docs](https://lrcmux.dev/docs).

## Self-hosting

**Requirements:**

- Go
- Redis

### From source

```sh
git clone https://github.com/f1nniboy/lrcmux
cd lrcmux
cp config.example.toml config.toml  # edit as needed
go run ./cmd/lrcmux -config config.toml
```

### Fly.io

```sh
fly launch --no-deploy
fly secrets set REDIS_URL=your-redis-url
fly deploy
```

## Configuration

See `config.example.toml` for all available options.
