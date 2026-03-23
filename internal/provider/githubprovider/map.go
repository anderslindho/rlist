package githubprovider

import (
	"strings"
	"time"

	gh "github.com/google/go-github/v66/github"

	"github.com/ades/rlist/internal/repo"
)

func toRepo(r *gh.Repository) *repo.Repo {
	if r == nil {
		return nil
	}
	owner := r.GetOwner().GetLogin()
	name := r.GetName()
	if owner == "" || name == "" {
		return nil
	}
	out := &repo.Repo{
		Provider:      "github",
		Owner:         owner,
		Name:          name,
		Visibility:    visibility(r),
		Description:   r.GetDescription(),
		DefaultBranch: r.GetDefaultBranch(),
		Language:      r.GetLanguage(),
		Archived:      r.GetArchived(),
		Fork:          r.GetFork(),
		Disabled:      r.GetDisabled(),
		Template:      r.GetIsTemplate(),
		Stars:         r.GetStargazersCount(),
		ForkCount:     r.GetForksCount(),
		OpenIssues:    r.GetOpenIssuesCount(),
		Homepage:      strings.TrimSpace(r.GetHomepage()),
		WebURL:        r.GetHTMLURL(),
		HTTPURL:       r.GetCloneURL(),
		SSHURL:        r.GetSSHURL(),
		Topics:        append([]string(nil), r.Topics...),
		SizeKB:        int(r.GetSize()),
		OpenIssuesNote: "(includes PRs)",
	}
	out.CreatedAt = timestamp(r.GetCreatedAt())
	out.UpdatedAt = timestamp(r.GetUpdatedAt())
	out.PushedAt = timestamp(r.GetPushedAt())
	if lic := r.GetLicense(); lic != nil {
		out.LicenseSPDX = formatLicense(lic)
	}
	if r.GetFork() {
		if p := r.GetParent(); p != nil {
			po := p.GetOwner().GetLogin()
			pn := p.GetName()
			if po != "" && pn != "" {
				out.ParentFullPath = po + "/" + pn
			}
		}
	}
	return out
}

func visibility(r *gh.Repository) string {
	if v := r.GetVisibility(); v != "" {
		return v
	}
	if r.GetPrivate() {
		return "private"
	}
	return "public"
}

func timestamp(t gh.Timestamp) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return t.Time
}

func formatLicense(lic *gh.License) string {
	if lic == nil {
		return ""
	}
	spdx := lic.GetSPDXID()
	if spdx == "" {
		spdx = lic.GetName()
	}
	if spdx == "" {
		return ""
	}
	if spdx == "NOASSERTION" {
		return "(license not asserted on GitHub)"
	}
	return spdx
}
