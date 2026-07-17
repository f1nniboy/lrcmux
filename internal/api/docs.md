**{{.AppName}}** is a lyrics aggregation API. It queries multiple providers, caches results, and returns the best available lyrics for a given track.

## Providers

| Provider | ID | Notes |
| -------- | -- | ----- |
{{range .Providers}}| {{if .URL}}**[{{.Name}}]({{.URL}})**{{else}}**{{.Name}}**{{end}} | `{{.ID}}` | {{.Desc}} |
{{end}}

## Sync levels

| Level | Description |
| ----- | ----------- |
{{range .Levels}}| `{{.String}}` | {{.Desc}} |
{{end}}

## Formats

The same result can be returned in any of these formats via the `format` query parameter. If the underlying lyrics do not meet a format's minimum sync level, the request fails with a `400`.

| Format | Content-Type | Min level | Notes |
| ------ | ------------ | --------- | -------- |
{{range .Formats}}| `{{.Name}}` | `{{.ContentType}}` | {{.MinLevel}} | {{.Desc}} |
{{end}}
## User agent

While it's not mandatory, we encourage you to include a `User-Agent` header identifying your app and a link to its repo (or a way to contact you if something goes wrong).

```
User-Agent: MyApp v1.0.0 (https://github.com/example/myapp)
```
{{if .RateLimit}}

## Rate limiting

Up to **{{.RateLimit.Limit}}** requests per **{{.RateLimit.Window}}** window. Cache hits are free and do not count against the limit.

You should respect the `Retry-After` header on 429 responses.
{{end}}
