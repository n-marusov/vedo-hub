# Lefthook Reference

> Source: https://lefthook.dev/, https://github.com/evilmartians/lefthook
> Created: 2026-07-25
> Updated: 2026-07-25

## Overview

Lefthook is a Git hooks manager written in Go. It is fast (parallel execution), powerful (flexible file filtering, script execution, Docker support), and simple (single dependency-free binary). It works with Node.js, Ruby, Python, Go, and many other project types. Version 2.1.10 is the latest.

## Core Concepts

- **Hook**: A Git hook (pre-commit, pre-push, commit-msg, etc.) configured in `lefthook.yml`. Each hook groups commands and scripts.
- **Command**: An inline shell command defined in the config under `run:`.
- **Script**: An external executable file stored in `<source_dir>/<hook-name>/` and referenced by name in the config.
- **Jobs**: A unified array format (`jobs:`) introduced in v1.10.0 that combines commands and scripts in a single list with identical configuration options.
- **Config file**: `lefthook.yml` (or `.lefthook.yml`, `lefthook.yaml`, `lefthook.toml`, `lefthook.json`, `lefthook.jsonc`) — the primary configuration.
- **Local config**: `lefthook-local.yml` (or `.lefthook-local.yml`) — overrides main config, useful for personal overrides without affecting teammates. Can be used standalone.
- **File templates**: Placeholders in `run:` that get substituted with actual file lists: `{staged_files}`, `{push_files}`, `{all_files}`, `{files}`, `{cmd}`, `{0}..{N}` (git hook arguments), `{lefthook_job_name}`.

## Installation

### Method 1: Package managers (recommended)

**Node.js (npm/yarn/pnpm):**
```bash
npm install --save-dev lefthook
yarn add --dev lefthook
pnpm add -D lefthook
```

**Ruby (Gemfile):**
```ruby
group :development do
  gem "lefthook", require: false
end
```
Or globally: `gem install lefthook`

**Python:**
```bash
pipx install lefthook
```

**Go:**
```bash
go install github.com/evilmartians/lefthook/v2@v2.1.10
# or as a Go tool:
go get -tool github.com/evilmartians/lefthook/v2@v2.1.10
```

**Homebrew (macOS/Linux):**
```bash
brew install lefthook
```

**Windows (winget):**
```bash
winget install evilmartians.lefthook
```

**Debian/Ubuntu (APT):**
```bash
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.deb.sh' | sudo -E bash
sudo apt install lefthook
```

**RPM-based (CentOS/Fedora):**
```bash
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.rpm.sh' | sudo -E bash
sudo yum install lefthook
```

**Alpine (APK):**
```bash
sudo apk add --no-cache bash curl
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.alpine.sh' | sudo -E bash
sudo apk add lefthook
```

### Method 2: Standalone binary

Download from [GitHub Releases](https://github.com/evilmartians/lefthook/releases/latest), place in `$PATH`, then:
```bash
lefthook self-update
```

## Quick Start

```bash
# 1. Create lefthook.yml with hook configuration
# 2. Install hooks
lefthook install

# 3. Commit — hooks run automatically
git add -A && git commit -m "..."
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `lefthook install` | Create/update Git hooks based on config. Creates empty `lefthook.yml` if none exists. |
| `lefthook install -f` | Force reinstall hooks without sync info check |
| `lefthook install -a` | Clear `.git/hooks/` dir and reinstall hooks (aggressive) |
| `lefthook run <hook-name>` | Run a specific hook or group of commands |
| `lefthook run <hook-name> --jobs <job-names>` | Run only specific jobs within a hook |
| `lefthook validate` | Validate the configuration file |
| `lefthook dump` | Dump merged config (useful with remotes/extends) |
| `lefthook self-update` | Update lefthook binary to latest version |
| `lefthook check-install` | Check if hooks are installed properly |
| `lefthook add -d <hook-name>` | Create a script skeleton for a hook |
| `lefthook uninstall` | Remove lefthook hooks |

### Skip lefthook for a single commit:
```bash
LEFTHOOK=0 git commit
```

## Configuration Reference

### Config File Names

| Format | Acceptable names |
|--------|------------------|
| YAML | `lefthook.yml`, `lefthook.yaml`, `.lefthook.yml`, `.lefthook.yaml`, `.config/lefthook.yml`, `.config/lefthook.yaml` |
| TOML | `lefthook.toml`, `.lefthook.toml`, `.config/lefthook.toml` |
| JSON | `lefthook.json`, `.lefthook.json`, `.config/lefthook.json` |
| JSONC | `lefthook.jsonc`, `.lefthook.jsonc`, `.config/lefthook.jsonc` |

### Merge Order

Settings are applied in this order (later overrides earlier):
1. `lefthook.yml` / main config file
2. `extends` — additional config files
3. `remotes` — remote configs from Git repositories
4. `lefthook-local.yml` — local user config (highest priority)

### Global (top-level) settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `min_version` | string | — | Minimum required lefthook version, fails if binary is older |
| `assert_lefthook_installed` | bool | false | Fail if lefthook executable can't be found in PATH |
| `colors` | bool/object | auto | Color output: `false`, `true`, or custom color map |
| `no_tty` | bool | false | Hide spinner and interactive output |
| `rc` | string | — | Path to an rc file (sh script) for ENV setup (e.g., nvm) |
| `source_dir` | string | `.lefthook/` | Directory for hook scripts |
| `extends` | string/array | — | Path(s) to additional config files to merge (supports globs) |
| `remotes` | array | — | Remote Git repos with shared configs |
| `glob_matcher` | string | gobwas | Glob engine: `gobwas` or `doublestar` |
| `output` | array/all/false | all | Control verbosity: `meta`, `summary`, `empty_summary`, `success`, `failure`, `execution`, `execution_out`, `execution_info`, `skips` |

### Hook-level settings

Each hook name (e.g., `pre-commit`, `pre-push`, `commit-msg`, `post-checkout`) is a section in the config:

```yaml
pre-commit:
  parallel: true          # Run commands in parallel (default: false)
  piped: false            # Stop on first failure (default: false)
  follow: false           # Stream STDOUT in real-time (default: false)
  fail_on_changes: never  # Fail if files modified: never/always/ci/non-ci
  no_tty: false           # Override global no_tty for this hook
  skip: [...]             # Skip conditions (merge, rebase, ref, run)
  only: [...]             # Only run conditions (opposite of skip)
  exclude_tags: [...]     # Exclude commands/scripts by tag or name
  files: "git command"    # Custom file listing command (hook-level)
  exclude: [...]          # Glob patterns to exclude files (hook-level)
  source_dir: ".lefthook/" # Override global source_dir
  commands:               # Map of named inline commands (legacy format)
    <name>:
      run: "..."
  scripts:                # Map of named scripts (legacy format)
    "<filename>":
      runner: bash
  jobs:                   # Unified jobs array (v1.10+, preferred)
    - name: <string>
      run: <string>
      script: <string>
```

### Job/Command-level settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `name` | string | — | Job name for output and selective run |
| `run` | string | — | Shell command to execute (supports file templates) |
| `script` | string | — | Script filename (without path, from source_dir/<hook>/) |
| `runner` | string | — | Runner for scripts (e.g., `bash`, `node`, `docker run ... {cmd}`) |
| `glob` | string/array | — | Glob pattern(s) to filter files |
| `exclude` | array | — | Glob patterns to exclude files |
| `file_types` | string/array | — | Filter by file type (text, binary, executable, symlink, MIME types) |
| `files` | string | — | Custom command to list files for `{files}` template |
| `root` | string | — | Change working directory for the command |
| `priority` | int | 0 | Execution order (1=first, 0=last, only with piped or non-parallel) |
| `tags` | array | — | Tags for grouping/excluding |
| `env` | map | — | Environment variables for the command |
| `skip` | bool/array | false | Skip conditions: `true`, `merge`, `rebase`, `merge-commit`, `ref: <branch>`, `run: <cmd>` |
| `only` | bool/array | — | Only-run conditions (same values as skip) |
| `interactive` | bool | false | Enable interactive mode (opens /dev/tty) |
| `use_stdin` | bool | false | Pass stdin from OS to command (needed for `pre-push` stdin scripts) |
| `stage_fixed` | bool | false | Auto `git add` modified files after run (pre-commit only) |
| `fail_text` | string | — | Custom error message on failure |
| `timeout` | duration | — | Command timeout (e.g., `30s`, `5m`) |

### File Templates in `run`

| Template | Description | Available in |
|----------|-------------|--------------|
| `{staged_files}` | Files staged for commit | pre-commit |
| `{push_files}` | Files that are committed but not pushed | pre-push |
| `{all_files}` | All files tracked by git | any hook |
| `{files}` | Result of custom `files:` command | any hook |
| `{cmd}` | The raw command string from config | any |
| `{0}` | Single space-joined git hook arguments | any |
| `{1}` | 1st git hook argument | any |
| `{lefthook_job_name}` | Current job/command/script name | any |

### Output Control

```yaml
output:
  - meta           # Print lefthook version
  - summary        # Print summary block (successful and failed steps)
  - empty_summary  # Print summary heading when no steps to run
  - success        # Print successful steps
  - failure        # Print failed steps
  - execution      # Print execution logs
  - execution_out  # Print execution output
  - execution_info # Print "EXECUTE > ..." logging
  - skips          # Print "skip" messages
```

Override with env: `LEFTHOOK_OUTPUT="meta,success,summary"`

### Skip/Only Conditions

| Value | Condition |
|-------|-----------|
| `true` | Always skip |
| `merge` | Skip during merge conflict resolution |
| `rebase` | Skip during rebase |
| `merge-commit` | Skip when HEAD is a merge commit |
| `ref: main` | Skip when on branch `main` (supports globs: `ref: dev/*`) |
| `run: test ...` | Skip when shell test returns exit code 0 |

### Remotes (shared configs)

```yaml
remotes:
  - git_url: git@github.com:evilmartians/lefthook
    ref: v1.0.0
    configs:
      - examples/ruby-linter.yml
```

### Colors Configuration

```yaml
colors:
  cyan: 14
  gray: 244
  green: '#32CD32'
  red: '#FF1493'
  yellow: '#F0E68C'
```

## Usage Patterns

### Basic lint-on-commit

```yaml
# lefthook.yml
pre-commit:
  parallel: true
  jobs:
    - run: yarn eslint --fix '{staged_files}'
      glob: "*.{js,ts,jsx,tsx}"
      stage_fixed: true
    - run: yarn stylelint --fix '{staged_files}'
      glob: "*.css"
      stage_fixed: true
```

### Multiple file filter types

```yaml
pre-commit:
  jobs:
    - run: yarn lint {staged_files}
      file_types: text
    - run: yarn check-hex {staged_files}
      file_types: binary
```

### Using scripts

```yaml
commit-msg:
  scripts:
    "template_checker":
      runner: bash
```

Scripts stored in `.lefthook/commit-msg/template_checker`.

### Priority (sequential ordering)

```yaml
post-checkout:
  piped: true
  jobs:
    - name: db-create
      priority: 1
      run: rails db:create
    - name: db-migrate
      priority: 2
      run: rails db:migrate
    - name: db-seed
      priority: 3
      run: rails db:seed
```

### Custom file list

```yaml
pre-push:
  jobs:
    - name: stylelint
      files: git diff --name-only master
      glob: "*.js"
      run: yarn stylelint {files}
```

### Environment variables

```yaml
pre-commit:
  jobs:
    - name: test
      env:
        RAILS_ENV: test
      run: bundle exec rspec
```

### Docker integration

```yaml
# lefthook.yml
pre-commit:
  jobs:
    - script: "good_job.js"
      runner: docker run -it --rm <container_id_or_name> {cmd}
```

### Local config overrides

```yaml
# lefthook-local.yml (not committed)
pre-push:
  exclude_tags:
    - frontend
  jobs:
    - name: audit packages
      skip: true
```

### Custom tasks (non-hook groups)

```yaml
# lefthook.yml
fixer:
  jobs:
    - run: bundle exec rubocop --safe-auto-correct -- {staged_files}
    - run: yarn eslint --fix {staged_files}
```

```bash
lefthook run fixer
```

### Fail on changes for CI

```yaml
pre-commit:
  parallel: true
  fail_on_changes: "ci"   # fail only in CI when files were modified
  jobs:
    - run: yarn lint --fix {staged_files}
      glob: "*.{js,ts}"
```

### Using rc file for ENV (nvm/fnm)

```yaml
# lefthook-local.yml
rc: ~/.lefthookrc
```

```bash
# ~/.lefthookrc
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
PATH=$PATH:$HOME/.nvm/versions/node/v15.14.0/bin
```

## Best Practices

1. **Always use `parallel: true`** for lint commands to maximize speed — Lefthook's main advantage.
2. **Use `stage_fixed: true`** with auto-fix linters so fixes are automatically staged (pre-commit only).
3. **Use `jobs` array format** (v1.10+) instead of separate `commands`/`scripts` maps for consistency.
4. **Place personal overrides in `lefthook-local.yml`** and add it to `.gitignore` so teammates are not affected.
5. **Use `fail_on_changes: ci`** to allow local auto-fixes while enforcing clean commits in CI.
6. **Use `tags` with `exclude_tags`** in local config to selectively skip groups of commands.
7. **Set `min_version`** to ensure consistent behavior across team members.
8. **Use `rc` file** when hooks run from GUI tools (VSCode) that may not have nvm/fnm in PATH.
9. **Prefer `{staged_files}`** over `{all_files}` in pre-commit to only lint changed files.
10. **Add `lefthook` to `pnpm.onlyBuiltDependencies`** in `package.json` when using pnpm.

## Common Pitfalls

- **`**` matches 1+ directories** with the default `gobwas` globber. Use `glob_matcher: doublestar` for standard behavior where `**` matches 0+ directories.
- **Command line length limits** — Lefthook automatically batches file lists, but very large projects may need attention.
- **`piped` and `parallel` are mutually exclusive** — setting both will cause an error.
- **pnpm users** must configure `onlyBuiltDependencies` in pnpm-workspace.yaml and package.json, otherwise the postinstall script won't run.
- **Globs ignore `root` option** — globs are always calculated from the Git repo root, not from `root:`.
- **GUI tools (VSCode) may not find nvm/fnm executables** — use the `rc` option to source your shell config.
- **`lefthook-local.yml` with leading dot** — if your main config uses a leading dot (`.lefthook.yml`), the local config must also use a leading dot (`.lefthook-local.yml`).

## Environment Variables

| Variable | Description |
|----------|-------------|
| `LEFTHOOK=0` | Skip lefthook for a single Git command |
| `LEFTHOOK_VERBOSE` | Enable verbose output |
| `LEFTHOOK_QUIET` | Suppress non-error output |
| `LEFTHOOK_OUTPUT` | Override `output` config (comma-separated values) |
| `LEFTHOOK_EXCLUDE` | Override `exclude_tags` config |
| `LEFTHOOK_BIN` | Custom path to lefthook executable |
| `LEFTHOOK_CONFIG` | Custom path to config file |
| `NO_COLOR` | Disable colored output |
| `CLICOLOR_FORCE` | Force colored output |
| `CI` | Used by `fail_on_changes: ci` |

## Version Notes

| Version | Key Changes |
|---------|-------------|
| **2.1.10** | AI coding agents integration; latest stable |
| **2.1.0** | `core.hooksPath` check; non-git hook installation |
| **2.0.0** | Breaking: `exclude` regexp support dropped (globs only); `skip_output` removed (use `output`); CLI args renamed for consistency; `sh` as command executor on Windows |
| **1.13.0** | `fail_on_changes` option |
| **1.12.0** | Install only specific hooks; folder restructuring |
| **1.10.0** | `jobs` array format introduced (unified commands + scripts) |
| **1.9.0** | Replaced viper with koanf (breaking for some edge cases) |
| **1.8.0** | Breaking: don't auto-install with npx; breaking: `files` command executed within configured `root` |
| **1.7.6** | `self-update` command added |
| **1.6.0** | `remotes` and `configs` options for shared configs |
| **1.4.0** | `dump` command, adaptive colors |
| **1.3.0** | Staged files change awareness |
| **1.0.0** | First stable major release |
