package provider

import (
	"fmt"
	"strings"

	"github.com/ades/rlist/internal/cache"
	"github.com/ades/rlist/internal/config"
	"github.com/ades/rlist/internal/provider/githubprovider"
	"github.com/ades/rlist/internal/provider/gitlabprovider"
)

// Resolve constructs the named provider (github or gitlab).
func Resolve(cfg config.Config, name string) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "github":
		return githubprovider.New()
	case "gitlab":
		return gitlabprovider.New(cfg)
	default:
		return nil, fmt.Errorf("unknown provider %q (use github or gitlab)", name)
	}
}

// ResolveDefault uses EffectiveProviderName with flag and config.
func ResolveDefault(cfg config.Config, providerFlag string) (Provider, error) {
	name := config.EffectiveProviderName(providerFlag, cfg)
	return Resolve(cfg, name)
}

// FromCache returns the provider that produced the cached row; host must match current config.
func FromCache(cfg config.Config, e cache.Entry) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(e.Provider)) {
	case "github":
		if e.Host != "" && e.Host != "github.com" {
			return nil, fmt.Errorf("cache entry host %q does not match GitHub", e.Host)
		}
		return githubprovider.New()
	case "gitlab":
		gl, err := gitlabprovider.New(cfg)
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(strings.TrimSpace(e.Host), gl.HostKey()) {
			return nil, fmt.Errorf("cache entry is for GitLab host %q but current base URL is %q (host %q); run rlist ls again for this instance",
				e.Host, cfg.EffectiveGitLabBaseURL(), gl.HostKey())
		}
		return gl, nil
	default:
		return nil, fmt.Errorf("unknown cached provider %q", e.Provider)
	}
}
