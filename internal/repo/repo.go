package repo

import (
	"strings"
	"time"
)

// Repo is a provider-neutral view used for list, show, and JSON output.
type Repo struct {
	Provider string

	Owner string
	Name  string

	Visibility string

	Description   string
	DefaultBranch string
	Language      string

	Archived bool
	Fork     bool
	Disabled bool
	Template bool

	ParentFullPath string

	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt  time.Time

	Stars      int
	ForkCount  int
	OpenIssues int

	Homepage string
	WebURL   string
	HTTPURL  string
	SSHURL   string

	LicenseSPDX string
	Topics      []string
	SizeKB      int

	OpenIssuesNote string
}

// SplitOwnerName splits path_with_namespace (e.g. "a/b/c") into owner "a/b" and name "c".
func SplitOwnerName(pathWithNamespace string) (owner, name string) {
	pathWithNamespace = strings.Trim(pathWithNamespace, "/")
	i := strings.LastIndex(pathWithNamespace, "/")
	if i < 0 {
		return "", pathWithNamespace
	}
	return pathWithNamespace[:i], pathWithNamespace[i+1:]
}
