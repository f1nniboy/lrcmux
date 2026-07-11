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
	"github.com/f1nniboy/lrcmux/internal/providers"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"
)

type docsData struct {
	RateLimit *rateLimitDoc
	AppName   string
	AppDomain string
	Levels    []lyrics.SyncLevel
	Formats   []formatEntry
	Providers []providers.Provider
}

// templates can't call multi-return funcs or access the registry name,
// so we need this wrapper
type formatEntry struct {
	format.Encoder
	Name string
}

func (f formatEntry) MinLevel() string { lo, _ := f.Encoder.Levels(); return lo.String() }

type rateLimitDoc struct {
	Window string
	Limit  int
}

func renderDocs(tmpl string, orch *orchestrator.Orchestrator, rate *ratelimit.Limiter, hide bool) (string, error) {
	t, err := template.New("docs").Parse(tmpl)
	if err != nil {
		return "", err
	}

	d := docsData{
		AppName:   meta.AppName,
		AppDomain: meta.AppDomain,
		Levels:    lyrics.Levels,
	}

	for _, name := range format.All() {
		enc, _ := format.Get(name)
		d.Formats = append(d.Formats, formatEntry{Encoder: enc, Name: name})
	}

	if !hide {
		d.Providers = orch.Providers()
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
