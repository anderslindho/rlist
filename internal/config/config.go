package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	EnvDefaultProvider = "RLIST_PROVIDER"
	envLegacyProvider  = "GHG_PROVIDER"
	EnvGitLabBaseURL   = "GITLAB_BASE_URL"
)

const configFileName = ".rlistrc"

// Config holds non-secret settings from file and environment.
type Config struct {
	DefaultProvider string `yaml:"default_provider"`
	GitLab          struct {
		BaseURL string `yaml:"base_url"`
	} `yaml:"gitlab"`
}

// Load reads ~/.rlistrc (YAML) in the user's home directory. Missing file is not an error. Environment overrides apply after merge.
func Load() (Config, error) {
	var c Config
	c.DefaultProvider = "github"

	path, err := configPath()
	if err != nil {
		return c, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			applyEnvOverrides(&c)
			return c, nil
		}
		return c, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("parse config %s: %w", path, err)
	}
	applyEnvOverrides(&c)
	return c, nil
}

// FilePath returns the path Load reads.
func FilePath() (string, error) {
	return configPath()
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home: %w", err)
	}
	return filepath.Join(home, configFileName), nil
}

func applyEnvOverrides(c *Config) {
	if v := strings.TrimSpace(os.Getenv(EnvDefaultProvider)); v != "" {
		c.DefaultProvider = v
	} else if v := strings.TrimSpace(os.Getenv(envLegacyProvider)); v != "" {
		c.DefaultProvider = v
	}
	if v := strings.TrimSpace(os.Getenv(EnvGitLabBaseURL)); v != "" {
		c.GitLab.BaseURL = v
	}
}

// EffectiveGitLabBaseURL returns the API base: env GITLAB_BASE_URL, else config.
func (c *Config) EffectiveGitLabBaseURL() string {
	if v := strings.TrimSpace(os.Getenv(EnvGitLabBaseURL)); v != "" {
		return v
	}
	return strings.TrimSpace(c.GitLab.BaseURL)
}

// GitLabHostKey returns the host part of the configured GitLab base URL (for cache validation).
func (c *Config) GitLabHostKey() (string, error) {
	raw := c.EffectiveGitLabBaseURL()
	if raw == "" {
		return "", fmt.Errorf("GitLab base URL is not set (set %s or gitlab.base_url in config)", EnvGitLabBaseURL)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse GitLab base URL: %w", err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("GitLab base URL %q has no host", raw)
	}
	return u.Host, nil
}

// EffectiveProviderName resolves CLI flag > RLIST_PROVIDER (or legacy GHG_PROVIDER) > config default > github.
func EffectiveProviderName(flag string, c Config) string {
	if v := strings.TrimSpace(flag); v != "" {
		return strings.ToLower(v)
	}
	if v := strings.TrimSpace(os.Getenv(EnvDefaultProvider)); v != "" {
		return strings.ToLower(v)
	}
	if v := strings.TrimSpace(c.DefaultProvider); v != "" {
		return strings.ToLower(v)
	}
	return "github"
}
