package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ades/rlist/internal/cache"
	"github.com/ades/rlist/internal/provider"
	"github.com/ades/rlist/internal/repo"
	"github.com/ades/rlist/internal/style"
)

// ListOptions filters and output mode for rlist ls.
type ListOptions struct {
	Mine       bool
	Org        string
	NoArchived bool
	OnlyForks  bool
	NoForks    bool
	JSON       bool
}

type repoListItem struct {
	Owner    string
	Name     string
	Vis      string
	Updated  time.Time
	Fork     bool
	Archived bool
}

type repoJSON struct {
	N          int    `json:"n"`
	Owner      string `json:"owner"`
	Name       string `json:"name"`
	FullName   string `json:"full_name"`
	Visibility string `json:"visibility"`
	Fork       bool   `json:"fork"`
	Archived   bool   `json:"archived"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

func List(ctx context.Context, p provider.Provider, opts ListOptions) error {
	if p == nil {
		return fmt.Errorf("provider is nil")
	}
	if opts.OnlyForks && opts.NoForks {
		return fmt.Errorf("--only-forks and --no-forks cannot be used together")
	}
	if opts.Mine && strings.TrimSpace(opts.Org) != "" {
		return fmt.Errorf("--mine and --org cannot be used together")
	}

	repos, err := p.ListRepos(ctx, repo.ListOpts{
		Mine:       opts.Mine,
		Org:        opts.Org,
		NoArchived: opts.NoArchived,
		OnlyForks:  opts.OnlyForks,
		NoForks:    opts.NoForks,
	})
	if err != nil {
		return err
	}

	items := reposToListItems(repos)
	items = applyListFilters(items, opts)

	sort.Slice(items, func(i, j int) bool {
		oi, oj := strings.ToLower(items[i].Owner), strings.ToLower(items[j].Owner)
		if oi != oj {
			return oi < oj
		}
		if !items[i].Updated.Equal(items[j].Updated) {
			return items[i].Updated.After(items[j].Updated)
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	host := p.HostKey()
	prov := p.Name()
	entries := make([]cache.Entry, len(items))
	for i := range items {
		entries[i] = cache.Entry{
			Provider: prov,
			Host:     host,
			Owner:    items[i].Owner,
			Name:     items[i].Name,
		}
	}

	if err := cache.Save(entries); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}

	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "no repositories")
		return nil
	}

	if opts.JSON {
		return writeListJSON(items)
	}

	st := style.NewStdout(true)
	return writeListTable(items, st)
}

func reposToListItems(repos []repo.Repo) []repoListItem {
	var items []repoListItem
	for _, r := range repos {
		if r.Owner == "" && r.Name == "" {
			continue
		}
		items = append(items, repoListItem{
			Owner:    r.Owner,
			Name:     r.Name,
			Vis:      r.Visibility,
			Updated:  r.UpdatedAt,
			Fork:     r.Fork,
			Archived: r.Archived,
		})
	}
	return items
}

func ownerMatchesListOrg(owner, org string) bool {
	if strings.EqualFold(owner, org) {
		return true
	}
	org = strings.Trim(strings.TrimSpace(org), "/")
	if org == "" {
		return false
	}
	return strings.HasPrefix(strings.ToLower(owner), strings.ToLower(org)+"/")
}

func applyListFilters(items []repoListItem, opts ListOptions) []repoListItem {
	var out []repoListItem
	orgFilter := strings.TrimSpace(opts.Org)
	for _, it := range items {
		if orgFilter != "" && !ownerMatchesListOrg(it.Owner, orgFilter) {
			continue
		}
		if opts.NoArchived && it.Archived {
			continue
		}
		if opts.OnlyForks && !it.Fork {
			continue
		}
		if opts.NoForks && it.Fork {
			continue
		}
		out = append(out, it)
	}
	return out
}

func writeListJSON(items []repoListItem) error {
	out := make([]repoJSON, len(items))
	for i, it := range items {
		updated := ""
		if !it.Updated.IsZero() {
			updated = it.Updated.UTC().Format(time.RFC3339)
		}
		full := it.Name
		if it.Owner != "" {
			full = it.Owner + "/" + it.Name
		}
		out[i] = repoJSON{
			N:          i + 1,
			Owner:      it.Owner,
			Name:       it.Name,
			FullName:   full,
			Visibility: it.Vis,
			Fork:       it.Fork,
			Archived:   it.Archived,
			UpdatedAt:  updated,
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func writeListTable(items []repoListItem, st *style.Styler) error {
	firstBlock := true
	start := 0
	for start < len(items) {
		owner := items[start].Owner
		end := start
		for end < len(items) && items[end].Owner == owner {
			end++
		}
		nInGroup := end - start

		if !firstBlock {
			fmt.Fprintln(os.Stdout)
		}
		firstBlock = false

		rule := fmt.Sprintf("── %s (%d) ──", owner, nInGroup)
		if st.Enabled {
			fmt.Fprintln(os.Stdout, st.Dim(rule))
		} else {
			fmt.Fprintln(os.Stdout, rule)
		}

		if err := writeListOwnerBlock(items, start, end, st); err != nil {
			return err
		}

		start = end
	}
	return nil
}

const listColGap = "  "

func writeListOwnerBlock(items []repoListItem, start, end int, st *style.Styler) error {
	wNum := len("#")
	for i := start; i < end; i++ {
		if n := len(strconv.Itoa(i + 1)); n > wNum {
			wNum = n
		}
	}
	wOwner, wRepo, wVis := len("OWNER"), len("REPO"), len("VIS")
	for i := start; i < end; i++ {
		it := items[i]
		if l := len(it.Owner); l > wOwner {
			wOwner = l
		}
		if l := len(it.Name); l > wRepo {
			wRepo = l
		}
		if l := len(it.Vis); l > wVis {
			wVis = l
		}
	}
	const wFork, wArch = 4, 4

	hdr := fmt.Sprintf("%-*s%s%-*s%s%-*s%s%-*s%s%-*s%s%-*s",
		wNum, "#", listColGap,
		wOwner, "OWNER", listColGap,
		wRepo, "REPO", listColGap,
		wVis, "VIS", listColGap,
		wFork, "fork", listColGap,
		wArch, "arch")
	if _, err := fmt.Fprintln(os.Stdout, hdr); err != nil {
		return err
	}

	for i := start; i < end; i++ {
		it := items[i]
		forkMark, archMark := "-", "-"
		if it.Fork {
			forkMark = "f"
		}
		if it.Archived {
			archMark = "a"
		}
		left := fmt.Sprintf("%-*d%s%-*s%s%-*s%s",
			wNum, i+1, listColGap,
			wOwner, it.Owner, listColGap,
			wRepo, it.Name, listColGap)
		if _, err := fmt.Fprint(os.Stdout, left); err != nil {
			return err
		}
		if _, err := fmt.Fprint(os.Stdout, formatVisPadded(st, it.Vis, wVis)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(os.Stdout, "%s%-*s%s%-*s\n", listColGap, wFork, forkMark, listColGap, wArch, archMark); err != nil {
			return err
		}
	}
	return nil
}

func formatVisPadded(st *style.Styler, plain string, w int) string {
	if len(plain) > w {
		w = len(plain)
	}
	pad := w - len(plain)
	if !st.Enabled {
		return fmt.Sprintf("%-*s", w, plain)
	}
	return st.Vis(plain) + strings.Repeat(" ", pad)
}
