package githubprovider

import (
	"fmt"
	"net/http"
	"os"
	"time"

	gh "github.com/google/go-github/v66/github"
)

func newGHClient() (*gh.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set (export it or add GITHUB_TOKEN=... to a .env in this directory or a parent); create a token at https://github.com/settings/tokens with repo scope for private repositories")
	}
	tr := cloneDefaultTransport()
	tr.MaxIdleConnsPerHost = 32
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Minute,
	}
	return gh.NewClient(httpClient).WithAuthToken(token), nil
}

func cloneDefaultTransport() *http.Transport {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		return t.Clone()
	}
	return &http.Transport{Proxy: http.ProxyFromEnvironment}
}
