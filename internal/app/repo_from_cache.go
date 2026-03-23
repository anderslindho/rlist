package app

import (
	"context"
	"fmt"

	"github.com/ades/rlist/internal/cache"
	"github.com/ades/rlist/internal/config"
	"github.com/ades/rlist/internal/provider"
	"github.com/ades/rlist/internal/repo"
)

func repositoryForListIndex(ctx context.Context, cfg config.Config, index int) (*repo.Repo, cache.Entry, error) {
	entries, err := cache.Load()
	if err != nil {
		return nil, cache.Entry{}, err
	}
	if index < 1 || index > len(entries) {
		return nil, cache.Entry{}, fmt.Errorf("index %d out of range (1–%d)", index, len(entries))
	}

	e := entries[index-1]
	p, err := provider.FromCache(cfg, e)
	if err != nil {
		return nil, e, err
	}

	r, err := p.GetRepo(ctx, e.Owner, e.Name)
	if err != nil {
		return nil, e, err
	}
	return r, e, nil
}
