**{{.AppName}}** is a lyrics aggregation API. It queries multiple providers in parallel, caches results, and returns the best available lyrics for a given track.
{{if .Providers}}

## Providers

| Provider | Max level | Notes |
| -------- | --------- | ----- |
{{range .Providers}}| **{{.Name}}** | {{.MaxLevel}} | {{.Desc}} |
{{end}}{{end}}

## Sync levels

| Level | Description |
| ----- | ----------- |
{{range .Levels}}| `{{.Name}}` | {{.Description}} |
{{end}}

## Formats

The same result can be returned in any of these formats via the `format` query parameter. If the underlying lyrics do not meet a format's minimum sync level, the request fails with a `400`.

| Format | Content-Type | Min level | Use case |
| ------ | ------------ | --------- | -------- |
{{range .Formats}}| `{{.Name}}` | `{{.ContentType}}` | {{.MinLevel}} | {{.UseCase}} |
{{end}}
## User agent

While it's not mandatory, we encourage you to include a `User-Agent` header identifying your app, its version, and a link to its homepage.

```
User-Agent: MyApp v1.0.0 (https://github.com/example/myapp)
```
{{if .RateLimit}}

## Rate limiting

Up to **{{.RateLimit.Limit}}** requests per **{{.RateLimit.Window}}** window (~{{.RateLimit.Rate}} sustained). Only live provider fetches count against the limit, cache hits are always free.
{{end}}
