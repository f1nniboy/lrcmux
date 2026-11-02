package api

import (
	"cmp"
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

type sourceInfo struct {
	lyrics.Source
	Level string `json:"level" enum:"word,line,none" doc:"Best sync level the source has"`
	Desc  string `json:"desc"`
}

type sourcesOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         []sourceInfo
}

func (s *Server) sourcesOp() huma.Operation {
	return huma.Operation{
		OperationID: "get-sources",
		Method:      http.MethodGet,
		Path:        "/sources",
		Summary:     "Lyrics sources",
		Description: "The sources lyrics come from. Their IDs are what the sources parameter takes.",
		Tags:        []string{"Meta"},
	}
}

func (s *Server) handleSources(context.Context, *struct{}) (*sourcesOutput, error) {
	provs := slices.Clone(s.orch.Providers())
	slices.SortStableFunc(provs, func(a, b providers.Provider) int { return cmp.Compare(b.MaxLevel(), a.MaxLevel()) })

	out := &sourcesOutput{CacheControl: "public, max-age=3600", Body: make([]sourceInfo, len(provs))}
	for i, p := range provs {
		out.Body[i] = sourceInfo{
			Source: providers.Source(p),
			Level:  p.MaxLevel().String(),
			Desc:   p.Desc(),
		}
	}
	return out, nil
}

type versionOutput struct {
	Body struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}
}

func (s *Server) versionOp() huma.Operation {
	return huma.Operation{
		OperationID: "get-version",
		Method:      http.MethodGet,
		Path:        "/version",
		Summary:     "Server version",
		Tags:        []string{"Meta"},
	}
}

func (s *Server) handleVersion(context.Context, *struct{}) (*versionOutput, error) {
	out := &versionOutput{}
	out.Body.Version = meta.Version
	out.Body.Commit = cmp.Or(meta.Commit, "unknown")
	return out, nil
}
