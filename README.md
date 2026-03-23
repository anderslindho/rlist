# rlist

Small CLI to **list**, **inspect**, and **open in a browser** repositories on **GitHub** or **GitLab** (including self-managed). The name is short for keeping your remotes in one place—without sounding tied to a single host.

## Install

From a clone of this repository:

```bash
go install ./cmd/rlist
```

Ensure `$(go env GOPATH)/bin` is on your `PATH` (often `~/go/bin`).

If the module is published at `github.com/ades/rlist`:

```bash
go install github.com/ades/rlist/cmd/rlist@latest
```

## Quickstart

1. **GitHub** — set a token with repo access:

   ```bash
   export GITHUB_TOKEN=ghp_...
   rlist ls
   ```

2. **GitLab** — set a token and your instance URL (non-secret settings can live in config):

   ```bash
   export GITLAB_TOKEN=glpat-...
   export GITLAB_BASE_URL=https://gitlab.example.com
   rlist ls --provider gitlab
   ```

   Or create `~/.rlistrc` (YAML; on Windows this is `%USERPROFILE%\.rlistrc`).

   ```yaml
   default_provider: gitlab
   gitlab:
     base_url: https://gitlab.example.com
   ```

3. **Details** — after `rlist ls`, row numbers are cached:

   ```bash
   rlist show 3
   rlist browse 3
   ```

4. Optional: put tokens in a `.env` file (current or parent directory); variables already set in the environment are not overwritten.

5. Default backend is `github`. Override with `--provider`, `RLIST_PROVIDER`, or `default_provider` in config. The legacy env var `GHG_PROVIDER` is still read if `RLIST_PROVIDER` is unset.

For every flag and behaviour, use:

```bash
rlist --help
rlist ls --help
```

## Paths

| Purpose   | Location |
| --------- | -------- |
| Config    | `~/.rlistrc` (YAML in your home directory; Windows: `%USERPROFILE%\.rlistrc`) |
| List cache | OS user cache dir, e.g. `~/Library/Caches/rlist/last.json` (macOS) |
