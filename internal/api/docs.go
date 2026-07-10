package api

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"
)

type docsData struct {
	RateLimit *rateLimitDoc
	AppName   string
	AppDomain string
	Levels    []levelDoc
	Formats   []formatDoc
	Providers []providerDoc
}

type providerDoc struct {
	Name string
	Desc string
}

type levelDoc struct {
	Name        string
	Description string
}

type formatDoc struct {
	Name        string
	ContentType string
	MinLevel    string
	UseCase     string
}

type rateLimitDoc struct {
	Window string
	Limit  int64
}

func renderDocs(tmpl string, orch *orchestrator.Orchestrator, rate *ratelimit.Limiter, hide bool) (string, error) {
	t, err := template.New("docs").Parse(tmpl)
	if err != nil {
		return "", err
	}

	d := docsData{
		AppName:   meta.AppName,
		AppDomain: meta.AppDomain,
	}

	for _, level := range lyrics.Levels {
		d.Levels = append(d.Levels, levelDoc{Name: level.String(), Description: level.Desc()})
	}

	for _, name := range format.All() {
		enc, _ := format.Get(name)
		lo, _ := enc.Levels()
		d.Formats = append(d.Formats, formatDoc{
			Name:        name,
			ContentType: enc.ContentType(),
			MinLevel:    lo.String(),
			UseCase:     enc.Desc(),
		})
	}

	if !hide {
		for _, p := range orch.Providers() {
			d.Providers = append(d.Providers, providerDoc{
				Name: p.Name(),
				Desc: p.Desc(),
			})
		}
	}

	if rate != nil {
		d.RateLimit = &rateLimitDoc{
			Limit:  rate.Limit(),
			Window: fmtDuration(rate.Window()),
		}
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, d); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func fmtDuration(d time.Duration) string {
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return d.String()
}
