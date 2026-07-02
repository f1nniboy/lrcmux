package api

import (
	"bytes"
	"fmt"
	"math"
	"text/template"
	"time"

	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"
)

type docsData struct {
	AppName   string
	AppDomain string
	Version   string
	Levels    []levelDoc
	Formats   []formatDoc
	Providers []providerDoc
	RateLimit *rateLimitDoc
}

type providerDoc struct {
	Name     string
	MaxLevel string
	Desc     string
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
	Limit  int64
	Window string
	Rate   string
}

func renderDocs(tmpl string, orch *orchestrator.Orchestrator, rl *ratelimit.Limiter, hide bool) (string, error) {
	t, err := template.New("docs").Parse(tmpl)
	if err != nil {
		return "", err
	}

	d := docsData{
		AppName:   meta.AppName,
		AppDomain: meta.AppDomain,
		Version:   meta.Version,
	}

	for _, level := range lyrics.Levels {
		d.Levels = append(d.Levels, levelDoc{Name: level.String(), Description: level.Desc()})
	}

	for _, name := range format.All() {
		enc, _ := format.Get(name)
		d.Formats = append(d.Formats, formatDoc{
			Name:        name,
			ContentType: enc.ContentType(),
			MinLevel:    enc.MinLevel().String(),
			UseCase:     enc.Desc(),
		})
	}

	if !hide {
		for _, p := range orch.Providers() {
			d.Providers = append(d.Providers, providerDoc{
				Name:     p.Name(),
				MaxLevel: p.MaxLevel().String(),
				Desc:     p.Desc(),
			})
		}
	}

	if rl != nil {
		rate := float64(rl.Limit()) / rl.Window().Seconds()
		var rateStr string
		if rate >= 1 {
			rateStr = fmt.Sprintf("%.4g/s", rate)
		} else {
			rateStr = fmt.Sprintf("1/%gs", math.Round(1/rate))
		}
		d.RateLimit = &rateLimitDoc{
			Limit:  rl.Limit(),
			Window: fmtDuration(rl.Window()),
			Rate:   rateStr,
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
