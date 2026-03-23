package provider

import (
	"context"

	"github.com/ades/rlist/internal/repo"
)

// Provider lists and fetches repositories for one host (GitHub.com or a GitLab instance).
type Provider interface {
	Name() string
	HostKey() string
	ListRepos(ctx context.Context, opts repo.ListOpts) ([]repo.Repo, error)
	GetRepo(ctx context.Context, owner, name string) (*repo.Repo, error)
}
