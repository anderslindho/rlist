package gitlabprovider

import (
	"context"
	"fmt"
	"os"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ades/rlist/internal/config"
	"github.com/ades/rlist/internal/repo"
)

// P implements GitLab API access for one instance (base URL from config / env).
type P struct {
	client *gitlab.Client
	host   string
}

// New builds a client using GITLAB_TOKEN and config.EffectiveGitLabBaseURL().
func New(cfg config.Config) (*P, error) {
	token := strings.TrimSpace(os.Getenv("GITLAB_TOKEN"))
	if token == "" {
		return nil, fmt.Errorf("GITLAB_TOKEN is not set (personal access token with read_api or api scope)")
	}
	base := cfg.EffectiveGitLabBaseURL()
	if base == "" {
		return nil, fmt.Errorf("GitLab base URL is not set (set %s or gitlab.base_url in config)", config.EnvGitLabBaseURL)
	}
	host, err := cfg.GitLabHostKey()
	if err != nil {
		return nil, err
	}
	c, err := gitlab.NewClient(token, gitlab.WithBaseURL(base))
	if err != nil {
		return nil, fmt.Errorf("gitlab client: %w", err)
	}
	return &P{client: c, host: host}, nil
}

func (p *P) Name() string    { return "gitlab" }
func (p *P) HostKey() string { return p.host }

func (p *P) ListRepos(ctx context.Context, opts repo.ListOpts) ([]repo.Repo, error) {
	order := "last_activity_at"
	sort := "desc"
	perPage := int64(100)

	if org := strings.TrimSpace(opts.Org); org != "" {
		base := gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{PerPage: perPage},
			OrderBy:     gitlab.Ptr(order),
			Sort:        gitlab.Ptr(sort),
		}
		if opts.NoArchived {
			base.Archived = gitlab.Ptr(false)
		}
		base.IncludeSubGroups = gitlab.Ptr(true)
		projects, err := listAllGroupProjects(ctx, p.client, org, base)
		if err != nil {
			return nil, err
		}
		return projectsToRepos(projects, opts)
	}

	base := gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{PerPage: perPage},
		OrderBy:     gitlab.Ptr(order),
		Sort:        gitlab.Ptr(sort),
		Statistics:  gitlab.Ptr(true),
	}
	if opts.Mine {
		base.Owned = gitlab.Ptr(true)
	} else {
		base.Membership = gitlab.Ptr(true)
	}
	if opts.NoArchived {
		base.Archived = gitlab.Ptr(false)
	}

	projects, err := listAllUserProjects(ctx, p.client, base)
	if err != nil {
		return nil, err
	}
	return projectsToRepos(projects, opts)
}

func projectsToRepos(projects []*gitlab.Project, opts repo.ListOpts) ([]repo.Repo, error) {
	var out []repo.Repo
	for _, p := range projects {
		if opts.NoArchived && p.Archived {
			continue
		}
		fork := p.ForkedFromProject != nil
		if opts.OnlyForks && !fork {
			continue
		}
		if opts.NoForks && fork {
			continue
		}
		if mr := toRepo(p); mr != nil {
			out = append(out, *mr)
		}
	}
	return out, nil
}

func (p *P) GetRepo(ctx context.Context, owner, name string) (*repo.Repo, error) {
	path := strings.TrimSpace(strings.Trim(owner, "/") + "/" + strings.Trim(name, "/"))
	path = strings.Trim(path, "/")
	if path == "" {
		return nil, fmt.Errorf("empty project path")
	}
	pid := path
	opt := &gitlab.GetProjectOptions{
		License:    gitlab.Ptr(true),
		Statistics: gitlab.Ptr(true),
	}
	proj, _, err := p.client.Projects.GetProject(pid, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	mr := toRepo(proj)
	if mr == nil {
		return nil, fmt.Errorf("get project: invalid response")
	}
	langs, _, err := p.client.Projects.GetProjectLanguages(pid, gitlab.WithContext(ctx))
	if err == nil {
		mergePrimaryLanguage(mr, langs)
	}
	return mr, nil
}

func listAllUserProjects(ctx context.Context, client *gitlab.Client, base gitlab.ListProjectsOptions) ([]*gitlab.Project, error) {
	var all []*gitlab.Project
	page := 1
	for {
		o := base
		o.Page = int64(page)
		projects, resp, err := client.Projects.ListProjects(&o, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("list projects: %w", err)
		}
		all = append(all, projects...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = int(resp.NextPage)
	}
	return all, nil
}

func listAllGroupProjects(ctx context.Context, client *gitlab.Client, gid string, base gitlab.ListGroupProjectsOptions) ([]*gitlab.Project, error) {
	var all []*gitlab.Project
	page := 1
	for {
		o := base
		o.Page = int64(page)
		projects, resp, err := client.Groups.ListGroupProjects(gid, &o, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("list group projects: %w", err)
		}
		all = append(all, projects...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = int(resp.NextPage)
	}
	return all, nil
}
