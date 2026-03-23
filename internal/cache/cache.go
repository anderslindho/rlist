package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const fileVersion = 2

// Entry identifies one row from the last list for show/browse.
type Entry struct {
	Provider string `json:"provider"`
	Host     string `json:"host"`
	Owner    string `json:"owner"`
	Name     string `json:"name"`
}

type file struct {
	Version int     `json:"version"`
	Repos   []Entry `json:"repos"`
}

func cacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache dir: %w", err)
	}
	return filepath.Join(base, "rlist"), nil
}

func cacheFilePath() (string, error) {
	d, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "last.json"), nil
}

func Save(repos []Entry) error {
	p, err := cacheFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	data, err := json.MarshalIndent(file{Version: fileVersion, Repos: repos}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write cache temp file: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		return fmt.Errorf("replace cache file: %w", err)
	}
	return nil
}

func Load() ([]Entry, error) {
	p, err := cacheFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("no cached list: run rlist ls first")
		}
		return nil, fmt.Errorf("read cache: %w", err)
	}
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("corrupt cache: %w", err)
	}
	if f.Version == 0 && len(f.Repos) > 0 {
		migrateV1(&f)
	}
	for i := range f.Repos {
		if err := validateEntry(&f.Repos[i]); err != nil {
			return nil, fmt.Errorf("corrupt cache: %w", err)
		}
	}
	return f.Repos, nil
}

func migrateV1(f *file) {
	for i := range f.Repos {
		if strings.TrimSpace(f.Repos[i].Provider) == "" {
			f.Repos[i].Provider = "github"
		}
		if strings.TrimSpace(f.Repos[i].Host) == "" {
			f.Repos[i].Host = "github.com"
		}
	}
	f.Version = fileVersion
}

func validateEntry(e *Entry) error {
	if strings.TrimSpace(e.Provider) == "" {
		return fmt.Errorf("missing provider on cache entry")
	}
	if strings.TrimSpace(e.Host) == "" {
		return fmt.Errorf("missing host on cache entry")
	}
	if strings.TrimSpace(e.Name) == "" {
		return fmt.Errorf("missing name on cache entry")
	}
	return nil
}
