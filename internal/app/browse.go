package app

import (
	"context"
	"fmt"

	"github.com/ades/rlist/internal/config"
	"github.com/pkg/browser"
)

func Browse(ctx context.Context, cfg config.Config, index int) error {
	r, e, err := repositoryForListIndex(ctx, cfg, index)
	if err != nil {
		return err
	}

	url := r.WebURL
	if url == "" {
		if e.Provider == "github" {
			url = fmt.Sprintf("https://github.com/%s/%s", e.Owner, e.Name)
		}
	}
	if url == "" {
		return fmt.Errorf("no web URL for this repository")
	}
	return browser.OpenURL(url)
}
