<p align="center">
  <img src="frontend/static/logo.svg" alt="lrcmux logo" width="128">
</p>

<h1 align="center">lrcmux</h1>

<p align="center">
  A lyrics aggregator API
  <br><br>
  <a href="https://github.com/f1nniboy/lrcmux/releases"><img src="https://img.shields.io/github/v/release/f1nniboy/lrcmux" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/f1nniboy/lrcmux?color=blue" alt="License"></a>
  <a href="https://matrix.to/#/#lrcmux:oss.zone"><img src="https://img.shields.io/matrix/lrcmux:oss.zone.svg?server_fqdn=matrix.oss.zone&fetchMode=summary&color=blue" alt="Matrix"></a>
</p>

Fans out requests across multiple providers, picks the best result, and caches everything.

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

### Fly.io + Cloudflare Workers

The API runs on Fly.io and the frontend on Cloudflare Workers. Set up the API first:

```sh
fly launch --no-deploy
fly secrets set REDIS_URL=redis://...
```

Then deploy both:

```sh
just deploy
```

## Configuration

See `config.example.toml` for all available options.

## I have a question!

Join the Matrix room at **[#lrcmux:oss.zone](https://matrix.to/#/#lrcmux:oss.zone)**.
