package githubprovider

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v66/github"

	"github.com/ades/rlist/internal/repo"
)

// P implements GitHub.com repository access.
type P struct {
	client *gh.Client
}

// New returns a Provider for github.com using GITHUB_TOKEN.
func New() (*P, error) {
	c, err := newGHClient()
	if err != nil {
		return nil, err
	}
	return &P{client: c}, nil
}

func (p *P) Name() string    { return "github" }
func (p *P) HostKey() string { return "github.com" }

func (p *P) ListRepos(ctx context.Context, opts repo.ListOpts) ([]repo.Repo, error) {
	var all []*gh.Repository
	var err error
	if org := strings.TrimSpace(opts.Org); org != "" {
		all, err = listReposForOwner(ctx, p.client, org)
	} else {
		repoType := "all"
		if opts.Mine {
			repoType = "owner"
		}
		repos, resp, err2 := p.client.Repositories.ListByAuthenticatedUser(ctx, authedRepoListOpts(repoType, 1))
		if err2 != nil {
			return nil, fmt.Errorf("list repositories: %w", err2)
		}
		all, err = continueFromFirstPage(ctx, repos, resp, func(ctx context.Context, page int) ([]*gh.Repository, *gh.Response, error) {
			return p.client.Repositories.ListByAuthenticatedUser(ctx, authedRepoListOpts(repoType, page))
		})
	}
	if err != nil {
		return nil, err
	}
	var out []repo.Repo
	for _, r := range all {
		if mr := toRepo(r); mr != nil {
			out = append(out, *mr)
		}
	}
	return out, nil
}

func (p *P) GetRepo(ctx context.Context, owner, name string) (*repo.Repo, error) {
	r, _, err := p.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}
	mr := toRepo(r)
	if mr == nil {
		return nil, fmt.Errorf("get repository: empty owner or name")
	}
	return mr, nil
}
