package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ades/rlist/internal/config"
	"github.com/ades/rlist/internal/style"
)

func Show(ctx context.Context, cfg config.Config, index int) error {
	r, e, err := repositoryForListIndex(ctx, cfg, index)
	if err != nil {
		return err
	}

	st := style.NewStdout(true)

	vis := r.Visibility
	desc := strings.TrimSpace(r.Description)
	if desc == "" {
		desc = "(no description)"
	}
	lang := r.Language
	if lang == "" {
		lang = "(none detected)"
	}

	title := e.Name
	if strings.TrimSpace(e.Owner) != "" {
		title = e.Owner + "/" + e.Name
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", st.Bold(title))

	kv := func(label, value string) {
		fmt.Fprintf(&b, "%s %s\n", st.Dim(label+":"), value)
	}
	ind := func(label, value string) {
		fmt.Fprintf(&b, "  %s %s\n", st.Dim(label+":"), value)
	}

	kv("visibility", st.Vis(vis))
	kv("language", lang)
	kv("default branch", nonEmpty(r.DefaultBranch, "—"))

	var status []string
	if r.Archived {
		status = append(status, st.Red("archived"))
	}
	if r.Fork {
		status = append(status, st.Yellow("fork"))
	}
	if r.Disabled {
		status = append(status, st.Red("disabled"))
	}
	if r.Template {
		status = append(status, st.Cyan("template"))
	}
	if len(status) > 0 {
		fmt.Fprintf(&b, "%s %s\n", st.Dim("status:"), strings.Join(status, "  "))
	}
	if r.Fork && r.ParentFullPath != "" {
		kv("parent", r.ParentFullPath)
	}

	fmt.Fprintf(&b, "\n%s\n", st.Dim("description"))
	fmt.Fprintf(&b, "%s\n\n", desc)

	fmt.Fprintf(&b, "%s\n", st.Dim("activity"))
	fmt.Fprintf(&b, "  created %s   pushed %s   updated %s\n",
		fmtDate(r.CreatedAt), fmtDate(r.PushedAt), fmtDate(r.UpdatedAt))
	note := r.OpenIssuesNote
	if note != "" {
		note = " " + st.Dim(note)
	}
	fmt.Fprintf(&b, "  stars %d   forks %d   open issues %d%s\n",
		r.Stars, r.ForkCount, r.OpenIssues, note)

	fmt.Fprintf(&b, "\n%s\n", st.Dim("links"))
	if hp := strings.TrimSpace(r.Homepage); hp != "" {
		ind("homepage", hp)
	}
	if u := r.WebURL; u != "" {
		ind("web", u)
	}
	if u := r.HTTPURL; u != "" {
		ind("clone", u)
	}
	if u := r.SSHURL; u != "" {
		ind("ssh", u)
	}

	lic := r.LicenseSPDX
	topics := r.Topics
	size := r.SizeKB
	if lic != "" || len(topics) > 0 || size > 0 {
		fmt.Fprintf(&b, "\n%s\n", st.Dim("meta"))
		if lic != "" {
			ind("license", lic)
		}
		if len(topics) > 0 {
			ind("topics", strings.Join(topics, ", "))
		}
		if size > 0 {
			src := "provider metadata"
			if r.Provider == "github" {
				src = "GitHub metadata"
			} else if r.Provider == "gitlab" {
				src = "GitLab repository size"
			}
			ind("size", fmt.Sprintf("%d KB (%s)", size, src))
		}
	}

	_, err = os.Stdout.WriteString(b.String())
	return err
}
