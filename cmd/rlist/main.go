package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ades/rlist/internal/app"
	"github.com/ades/rlist/internal/config"
	"github.com/ades/rlist/internal/dotenv"
	"github.com/ades/rlist/internal/provider"
)

const cmdName = "rlist"

var (
	providerFlag   string
	listMine       bool
	listOrg        string
	listNoArchived bool
	listOnlyForks  bool
	listNoForks    bool
	listJSON       bool
)

func main() {
	dotenv.LoadFromAncestors()
	args := os.Args[1:]
	if len(args) == 1 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			args = []string{"show", args[0]}
		}
	}
	rootCmd.SetArgs(args)
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmdName, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   cmdName,
	Short: "List and inspect repositories on GitHub or GitLab",
	Long:  "",
	SilenceUsage:  true,
	SilenceErrors: true,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List repositories (show/browse use this cache)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		p, err := provider.ResolveDefault(cfg, providerFlag)
		if err != nil {
			return err
		}
		return app.List(cmd.Context(), p, app.ListOptions{
			Mine:       listMine,
			Org:        listOrg,
			NoArchived: listNoArchived,
			OnlyForks:  listOnlyForks,
			NoForks:    listNoForks,
			JSON:       listJSON,
		})
	},
}

var showCmd = &cobra.Command{
	Use:   "show INDEX",
	Short: "Details for a cached list row (run ls first)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseListIndex(args[0])
		if err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return app.Show(cmd.Context(), cfg, n)
	},
}

var browseCmd = &cobra.Command{
	Use:   "browse INDEX",
	Short: "Open a cached list row in the browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseListIndex(args[0])
		if err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return app.Browse(cmd.Context(), cfg, n)
	},
}

func parseListIndex(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid index %q", s)
	}
	if n < 1 {
		return 0, fmt.Errorf("index must be >= 1")
	}
	return n, nil
}

func init() {
	cfgPath, err := config.FilePath()
	if err != nil {
		cfgPath = "~/.rlistrc"
	}
	rootCmd.Long = fmt.Sprintf(`Uses each host's REST API. Set GITHUB_TOKEN and/or GITLAB_TOKEN; optional config in %s. See README for GitLab URL and providers.

Commands: ls, show, browse. A lone positive integer is shorthand for %s show N. Use %s ls -h for list flags.`, cfgPath, cmdName, cmdName)

	rootCmd.PersistentFlags().StringVar(&providerFlag, "provider", "", "github or gitlab (default from config / "+config.EnvDefaultProvider+")")

	lsCmd.Flags().BoolVarP(&listMine, "mine", "m", false, "only repositories under your user (exclude orgs)")
	lsCmd.Flags().StringVar(&listOrg, "org", "", "only repositories under this owner (user or organization)")
	lsCmd.Flags().BoolVar(&listNoArchived, "no-archived", false, "omit archived repositories")
	lsCmd.Flags().BoolVar(&listOnlyForks, "only-forks", false, "only forks")
	lsCmd.Flags().BoolVar(&listNoForks, "no-forks", false, "omit forks")
	lsCmd.Flags().BoolVar(&listJSON, "json", false, "print JSON (stable for scripts; disables table styling)")

	rootCmd.AddCommand(lsCmd, showCmd, browseCmd)
}
