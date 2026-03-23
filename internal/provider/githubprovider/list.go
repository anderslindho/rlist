package githubprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	gh "github.com/google/go-github/v66/github"
	"golang.org/x/sync/errgroup"
)

const listPageConcurrency = 8

func isNotFound(err error) bool {
	var e *gh.ErrorResponse
	return errors.As(err, &e) && e.Response != nil && e.Response.StatusCode == http.StatusNotFound
}

func continueFromFirstPage(ctx context.Context, first []*gh.Repository, resp *gh.Response, fetch func(context.Context, int) ([]*gh.Repository, *gh.Response, error)) ([]*gh.Repository, error) {
	out := append([]*gh.Repository(nil), first...)
	if resp == nil {
		return out, nil
	}
	if resp.LastPage >= 2 {
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(listPageConcurrency)
		batches := make([][]*gh.Repository, resp.LastPage-1)
		for page := 2; page <= resp.LastPage; page++ {
			page := page
			i := page - 2
			g.Go(func() error {
				repos, _, err := fetch(gctx, page)
				if err != nil {
					return err
				}
				batches[i] = repos
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
		for _, b := range batches {
			out = append(out, b...)
		}
		return out, nil
	}
	for resp.NextPage != 0 {
		repos, next, err := fetch(ctx, resp.NextPage)
		if err != nil {
			return nil, err
		}
		out = append(out, repos...)
		if next == nil {
			break
		}
		resp = next
	}
	return out, nil
}

func listReposForOwner(ctx context.Context, client *gh.Client, owner string) ([]*gh.Repository, error) {
	owner = strings.TrimSpace(owner)
	orgOpts := func(page int) *gh.RepositoryListByOrgOptions {
		return &gh.RepositoryListByOrgOptions{
			Sort:        "updated",
			Direction:   "desc",
			ListOptions: gh.ListOptions{Page: page, PerPage: 100},
		}
	}
	repos, resp, err := client.Repositories.ListByOrg(ctx, owner, orgOpts(1))
	if err != nil {
		if !isNotFound(err) {
			return nil, fmt.Errorf("list organization repositories: %w", err)
		}
		userOpts := func(page int) *gh.RepositoryListByUserOptions {
			return &gh.RepositoryListByUserOptions{
				Type:        "owner",
				Sort:        "updated",
				Direction:   "desc",
				ListOptions: gh.ListOptions{Page: page, PerPage: 100},
			}
		}
		repos, resp, err = client.Repositories.ListByUser(ctx, owner, userOpts(1))
		if err != nil {
			return nil, fmt.Errorf("list user repositories: %w", err)
		}
		return continueFromFirstPage(ctx, repos, resp, func(ctx context.Context, page int) ([]*gh.Repository, *gh.Response, error) {
			return client.Repositories.ListByUser(ctx, owner, userOpts(page))
		})
	}
	return continueFromFirstPage(ctx, repos, resp, func(ctx context.Context, page int) ([]*gh.Repository, *gh.Response, error) {
		return client.Repositories.ListByOrg(ctx, owner, orgOpts(page))
	})
}

func authedRepoListOpts(repoType string, page int) *gh.RepositoryListByAuthenticatedUserOptions {
	return &gh.RepositoryListByAuthenticatedUserOptions{
		Type:        repoType,
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: gh.ListOptions{Page: page, PerPage: 100},
	}
}
