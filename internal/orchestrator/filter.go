package orchestrator

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/providers"
)

var ErrInvalidSource = errors.New("invalid source")

// restricts list of providers to the given subset
//
// non-prefixed entries mean "keep only these",
// "!"-prefixed entries mean "drop these"
func filterBySources(provs []providers.Provider, sources []string) ([]providers.Provider, error) {
	if len(sources) == 0 {
		return provs, nil
	}

	known := make(map[string]struct{}, len(provs))
	for _, p := range provs {
		known[p.ID()] = struct{}{}
	}

	var include, exclude []string
	for _, s := range sources {
		s = strings.ToLower(strings.TrimSpace(s))
		name, isExclude := strings.CutPrefix(s, "!")
		if name == "" {
			return nil, fmt.Errorf("%w: empty source name", ErrInvalidSource)
		}
		if _, ok := known[name]; !ok {
			return nil, fmt.Errorf("%w: unknown provider %q", ErrInvalidSource, name)
		}
		if isExclude {
			exclude = append(exclude, name)
		} else {
			include = append(include, name)
		}
	}

	if len(include) > 0 && len(exclude) > 0 {
		return nil, fmt.Errorf("%w: cannot mix include and exclude", ErrInvalidSource)
	}

	if len(include) > 0 {
		return slices.DeleteFunc(slices.Clone(provs), func(p providers.Provider) bool {
			return !slices.Contains(include, p.ID())
		}), nil
	}
	return slices.DeleteFunc(slices.Clone(provs), func(p providers.Provider) bool {
		return slices.Contains(exclude, p.ID())
	}), nil
}
