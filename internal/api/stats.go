package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
)

type statsOutput struct {
	Body struct {
		Version   string                      `json:"version"`
		Commit    string                      `json:"commit"`
		Providers []orchestrator.ProviderInfo `json:"providers,omitempty"`
	}
}

func (s *Server) statsOp() huma.Operation {
	return huma.Operation{
		OperationID: "get-stats",
		Method:      http.MethodGet,
		Path:        "/stats",
		Summary:     "Server statistics",
		Tags:        []string{"Meta"},
	}
}

func (s *Server) handleStats(ctx context.Context, _ *struct{}) (*statsOutput, error) {
	out := &statsOutput{}
	out.Body.Version = meta.Version
	if meta.Commit != "" {
		out.Body.Commit = meta.Commit
	} else {
		out.Body.Commit = "unknown"
	}
	if !s.hide {
		out.Body.Providers = s.orch.ProviderInfos(ctx)
	}
	return out, nil
}
