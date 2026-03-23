package gitlabprovider

import (
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ades/rlist/internal/repo"
)

func toRepo(p *gitlab.Project) *repo.Repo {
	if p == nil {
		return nil
	}
	path := strings.TrimSpace(p.PathWithNamespace)
	if path == "" {
		return nil
	}
	owner, name := repo.SplitOwnerName(path)
	if name == "" {
		return nil
	}
	out := &repo.Repo{
		Provider:       "gitlab",
		Owner:          owner,
		Name:           name,
		Visibility:     strings.ToLower(string(p.Visibility)),
		Description:    p.Description,
		DefaultBranch:  p.DefaultBranch,
		Archived:       p.Archived,
		Fork:           p.ForkedFromProject != nil,
		Disabled:       p.RepositoryAccessLevel == gitlab.DisabledAccessControl,
		Stars:          int(p.StarCount),
		ForkCount:      int(p.ForksCount),
		OpenIssues:     int(p.OpenIssuesCount),
		WebURL:         p.WebURL,
		HTTPURL:        p.HTTPURLToRepo,
		SSHURL:         p.SSHURLToRepo,
		Topics:         append([]string(nil), p.Topics...),
		OpenIssuesNote: "(GitLab issues; not GitHub PRs)",
	}
	if p.License != nil {
		k := strings.TrimSpace(p.License.Key)
		if k != "" {
			out.LicenseSPDX = k
		} else if n := strings.TrimSpace(p.License.Name); n != "" {
			out.LicenseSPDX = n
		}
	}
	if p.ForkedFromProject != nil && p.ForkedFromProject.PathWithNamespace != "" {
		out.ParentFullPath = p.ForkedFromProject.PathWithNamespace
	}
	if p.CreatedAt != nil {
		out.CreatedAt = *p.CreatedAt
	}
	if p.UpdatedAt != nil {
		out.UpdatedAt = *p.UpdatedAt
	}
	if p.LastActivityAt != nil {
		out.PushedAt = *p.LastActivityAt
		if out.UpdatedAt.IsZero() {
			out.UpdatedAt = *p.LastActivityAt
		}
	}
	if p.Statistics != nil && p.Statistics.RepositorySize > 0 {
		out.SizeKB = int(p.Statistics.RepositorySize / 1024)
	}
	return out
}

func mergePrimaryLanguage(r *repo.Repo, langs *gitlab.ProjectLanguages) {
	if r == nil || langs == nil || len(*langs) == 0 {
		return
	}
	var best string
	var max float32
	for k, v := range *langs {
		if v > max {
			max = v
			best = k
		}
	}
	if best != "" {
		r.Language = best
	}
}
