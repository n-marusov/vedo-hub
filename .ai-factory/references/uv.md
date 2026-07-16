# uv Reference

> Source: https://docs.astral.sh/uv/
> Created: 2026-07-16
> Updated: 2026-07-16

## Overview

uv is an extremely fast Python package and project manager, written in Rust, backed by Astral (the creators of Ruff). It serves as a single replacement for `pip`, `pip-tools`, `pipx`, `poetry`, `pyenv`, `twine`, `virtualenv`, and more. uv is 10-100x faster than `pip`, provides comprehensive project management with a universal lockfile, supports scripts with inline dependency metadata, installs and manages Python versions, runs tools published as Python packages, includes a pip-compatible interface, supports Cargo-style workspaces, and is disk-space efficient with a global cache. It can be installed without Rust or Python and supports macOS, Linux, and Windows.

## Core Concepts

- **Projects**: Python projects defined via `pyproject.toml`, managed with `uv init`, `uv add`, `uv remove`, `uv sync`, `uv lock`, `uv run`, and `uv build`.
- **Universal lockfile (`uv.lock`)**: Cross-platform TOML lockfile with exact resolved versions. Managed by uv, not edited manually. Should be checked into version control.
- **Scripts**: Standalone Python files (PEP 723) with inline dependency metadata. Run with `uv run <script.py>`.
- **Tools**: CLI tools provided by Python packages. Run ephemerally with `uvx` (alias for `uv tool run`) or install persistently with `uv tool install`.
- **Python management**: uv downloads and manages Python versions from the Astral `python-build-standalone` project. Automatic download on demand.
- **Workspaces**: Cargo-style multi-package projects sharing a single lockfile.
- **Virtual environments**: Created automatically in `.venv` within projects. Also managed with `uv venv`.
- **Global cache**: Deduplicated package cache at `$XDG_CACHE_HOME/uv` or `%LOCALAPPDATA%\uv\cache` on Windows.

## Commands

### `uv init` — Create a new project

```
uv init [OPTIONS] [PATH]
```

Creates a `pyproject.toml`, `.python-version`, `README.md`, `.gitignore`, and optionally `main.py`. Use `--lib` for library packages, `--app` for applications (default), `--script` for PEP 723 scripts, `--package` to make it distributable, `--build-backend` to choose a build backend (hatch, flit, pdm, poetry, setuptools, maturin, scikit, uv), `--vcs` to choose version control (git or none), `--python` to set the Python version, `--name` for project name.

### `uv add` — Add dependencies

```
uv add [OPTIONS] <PACKAGES|--requirements <REQUIREMENTS>>
```

Adds dependencies to `pyproject.toml`, updates lockfile and environment. Supports version constraints (`uv add 'requests==2.31.0'`), Git dependencies (`uv add git+https://github.com/psf/requests`), requirements files (`uv add -r requirements.txt`), dev dependencies (`--dev`), optional dependencies (`--optional`), and editable installs (`--editable`). Use `--script` to add to a PEP 723 script's inline metadata.

### `uv remove` — Remove dependencies

```
uv remove [OPTIONS] <PACKAGES>...
```

Removes dependencies from `pyproject.toml`, updates lockfile and environment.

### `uv sync` — Sync environment with lockfile

```
uv sync [OPTIONS]
```

Updates the project environment to match the lockfile. By default performs exact sync (removes extraneous packages). Use `--inexact` to keep extraneous packages. Supports `--frozen`, `--locked`, `--check`, `--dry-run`, `--all-extras`, `--all-packages`, `--no-install-project` for Docker layering.

### `uv lock` — Update lockfile

```
uv lock [OPTIONS]
```

Creates or updates `uv.lock`. Supports `--check`, `--check-exists` (`--frozen`), `--dry-run`, `--upgrade-package`, `--script` for locking PEP 723 scripts.

### `uv run` — Run commands in project environment

```
uv run [OPTIONS] [COMMAND]
```

Runs a command or script in the project environment. Automatically syncs environment before running. Supports `--with` (additional packages), `--with-requirements`, `--python`, `--no-project`, `--script`, `--module` (`-m`), `--frozen`, `--locked`, `--no-sync`, `--isolated`.

### `uv export` — Export lockfile

```
uv export [OPTIONS]
```

Exports `uv.lock` to `requirements.txt`, `pylock.toml` (PEP 751), or `cyclonedx1.5` formats. Use `--output-file` or `-o`.

### `uv tree` — Display dependency tree

```
uv tree [OPTIONS]
```

Shows the project's dependency tree. Supports `--depth`, `--invert`, `--outdated`, `--package`, `--universal`, `--show-sizes`, `--script`.

### `uv build` — Build distributions

```
uv build [OPTIONS] [SRC]
```

Builds source distributions and wheels into `dist/`. Supports `--sdist`, `--wheel`, `--all-packages` (workspace), `--package`.

### `uv publish` — Upload to index

```
uv publish [OPTIONS] [FILES]...
```

Uploads distributions to PyPI or custom index. Supports `--publish-url`, `--token`, `--username`/`--password`, `--check-url` for duplicate detection, `--trusted-publishing` for CI environments.

### `uv version` — Read/update project version

```
uv version [OPTIONS] [VALUE]
```

Shows or updates the project version. Supports `--bump` (major, minor, patch, stable, alpha, beta, rc, post, dev), `--short`, `--output-format json`.

### `uv format` — Format Python code

```
uv format [OPTIONS] [-- <EXTRA_ARGS>...]
```

Formats Python code using the Ruff formatter. Supports `--check`, `--diff`, `--version` (select Ruff version).

### `uv check` — Type check Python code

```
uv check [OPTIONS]
```

Type checks Python code using `ty`. Uses dependencies from the project environment.

### `uv audit` — Dependency vulnerability audit

```
uv audit [OPTIONS]
```

Audits dependencies for known vulnerabilities via OSV API. Supports `--ignore`, `--ignore-until-fixed`, `--output-format` (text, json, sarif).

### `uv tool` — Tool management

Subcommands:
- `uv tool run` / `uvx` — Run a tool ephemerally
- `uv tool install` — Install a tool persistently
- `uv tool upgrade` — Upgrade installed tools (or `--all`)
- `uv tool list` — List installed tools (supports `--outdated`, `--show-paths`)
- `uv tool uninstall` — Uninstall a tool
- `uv tool update-shell` — Add tool bin dir to PATH
- `uv tool dir` — Show tools directory path (`--bin` for executables)

### `uv python` — Python version management

Subcommands:
- `uv python install` — Download and install Python versions
- `uv python list` — List available/installed Python versions
- `uv python find` — Find a Python interpreter
- `uv python pin` — Pin Python version in `.python-version`
- `uv python dir` — Show Python installation directory (`--bin` for executables)
- `uv python uninstall` — Uninstall Python versions
- `uv python upgrade` — Upgrade to latest patch version
- `uv python update-shell` — Add Python bin dir to PATH

### `uv venv` — Create virtual environment

```
uv venv [OPTIONS] [PATH]
```

Creates a virtual environment (default `.venv`). Supports `--python`, `--seed`, `--allow-existing`, `--clear`, `--relocatable`, `--system-site-packages`, `--prompt`.

### `uv cache` — Cache management

Subcommands:
- `uv cache clean` — Clear cache (optionally for specific packages)
- `uv cache prune` — Remove dangling entries (`--ci` for CI optimization)
- `uv cache dir` — Show cache directory path
- `uv cache size` — Show cache size (`--human` for human-readable)

### `uv auth` — Authentication management

Subcommands:
- `uv auth login` — Login to a service
- `uv auth logout` — Logout from a service
- `uv auth token` — Show token for a service
- `uv auth dir` — Show credentials directory

### `uv self` — uv self-management

Subcommands:
- `uv self update` — Update uv itself
- `uv self version` — Show uv's version

### `uv pip` — Pip-compatible interface

Subcommands:
- `uv pip compile` — Compile requirements to locked format
- `uv pip sync` — Sync environment with requirements file
- `uv pip install` — Install packages
- `uv pip uninstall` — Uninstall packages
- `uv pip freeze` — List installed as requirements
- `uv pip list` — List installed in table format
- `uv pip show` — Show package info
- `uv pip tree` — Show dependency tree
- `uv pip check` — Verify installed dependencies

### `uv workspace` — Workspace inspection

Subcommands:
- `uv workspace metadata` — View workspace metadata
- `uv workspace dir` — Display workspace member path
- `uv workspace list` — List workspace members

## Installation

| Method | Command |
|--------|---------|
| Standalone (macOS/Linux) | `curl -LsSf https://astral.sh/uv/install.sh \| sh` |
| Standalone (Windows) | `powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 \| iex"` |
| pipx | `pipx install uv` |
| pip | `pip install uv` |
| Homebrew | `brew install uv` |
| WinGet | `winget install --id=astral-sh.uv -e` |
| Scoop | `scoop install main/uv` |
| Cargo | `cargo install --locked uv` |
| Docker | `ghcr.io/astral-sh/uv` |

Upgrade via standalone: `uv self update`. Self-updates are disabled for non-standalone installs.

## Key Environment Variables

| Variable | Purpose |
|----------|---------|
| `UV_PYTHON` | Default Python interpreter request |
| `UV_CACHE_DIR` | Custom cache directory |
| `UV_NO_CACHE` | Disable cache |
| `UV_OFFLINE` | Disable network access |
| `UV_FROZEN` | Don't update lockfile |
| `UV_LOCKED` | Assert lockfile is up-to-date |
| `UV_NO_SYNC` | Skip environment sync |
| `UV_NO_PROJECT` | Skip project discovery |
| `UV_BREAK_SYSTEM_PACKAGES` | Allow modifying system Python |
| `UV_SYSTEM_PYTHON` | Use system Python |
| `UV_COMPILE_BYTECODE` | Compile .py to bytecode |
| `UV_INDEX` | Custom package index URLs |
| `UV_DEFAULT_INDEX` | Default package index URL |
| `UV_NO_INDEX` | Ignore registry indexes |
| `UV_CONSTRAINT` | Constraint files |
| `UV_OVERRIDE` | Override files |
| `UV_EXCLUDE_NEWER` | Date-based cutoff |
| `UV_PRERELEASE` | Pre-release strategy |
| `UV_RESOLUTION` | Resolution strategy (highest/lowest/lowest-direct) |
| `UV_FORK_STRATEGY` | Multi-version fork strategy |
| `UV_INDEX_STRATEGY` | Multi-index strategy |
| `UV_KEYRING_PROVIDER` | Keyring authentication |
| `UV_LINK_MODE` | Package install method (clone/copy/hardlink/symlink) |
| `UV_PUBLISH_URL` | Publish endpoint URL |
| `UV_PUBLISH_TOKEN` | Publish token |
| `UV_PUBLISH_USERNAME` | Publish username |
| `UV_PUBLISH_PASSWORD` | Publish password |
| `UV_PUBLISH_CHECK_URL` | Duplicate upload check URL |
| `UV_WORKING_DIR` | Working directory override |
| `UV_PROJECT` | Project root directory |
| `UV_INSECURE_HOST` | Allow insecure hosts |
| `UV_NO_CONFIG` | Disable config file discovery |
| `UV_CONFIG_FILE` | Specific config file path |
| `UV_NO_PROGRESS` | Hide progress indicators |
| `UV_MANAGED_PYTHON` | Require uv-managed Python |
| `UV_NO_MANAGED_PYTHON` | Disable uv-managed Python |
| `UV_PYTHON_INSTALL_DIR` | Python installation directory |
| `UV_TOOL_DIR` | Tools directory |
| `UV_CREDENTIALS_DIR` | Credentials directory |
| `UV_GITHUB_TOKEN` | GitHub token for self-update |
| `UV_TORCH_BACKEND` | PyTorch CUDA/ROCm backend |

## Project Configuration (`pyproject.toml` `[tool.uv]`)

uv reads configuration from `pyproject.toml` (under `[tool.uv]`) and/or `uv.toml`. Key settings:

| Setting | Purpose |
|---------|---------|
| `dev-dependencies` | Dev-only dependencies |
| `sources` | Alternative sources (git, path, url) for dependencies |
| `index` | Additional package indexes |
| `default-index` | Default package index |
| `managed` | Whether uv manages the project |
| `workspace` | Workspace member configuration |
| `constraints` | Constraint files |
| `override` | Override files |
| `conflicts` | Conflicting extras/groups |
| `exclude-newer` | Date-based cutoff |
| `environments` | Environment markers for universal resolution |
| `default-groups` | Default dependency groups |
| `no-build-isolation` | Disable build isolation |
| `compile-bytecode` | Compile bytecode on install |
| `link-mode` | Package link mode |
| `python-downloads` | Automatic Python downloads |

## Python Version Request Formats

```
3              — latest stable Python 3
3.12           — latest patch of Python 3.12
3.12.3         — exact Python 3.12.3
>=3.12,<3.13   — version specifier
3.13t          — freethreaded (free-threaded build)
3.12.0d        — debug build
3.13+freethreaded — variant syntax
pypy           — latest PyPy
[email protected]   — PyPy 3.10
[email protected]   — PyPy 3.8 v7.3
cpython        — CPython (default)
cp312          — CPython 3.12 shorthand
/path/to/python — specific executable
python3.11     — specific executable name
```

## Best Practices

1. **Use `uv.lock` in version control** — ensures reproducible installations across machines.
2. **Use `uv sync` for CI/CD** — fast, deterministic environment setup. Use `--no-install-project` for Docker layer caching.
3. **Pin Python versions** — use `.python-version` files or `uv python pin` for project consistency.
4. **Use `uvx` for ephemeral tool usage** — avoids polluting project environments. Use `uv tool install` for frequently used tools.
5. **Declare script dependencies inline** — use PEP 723 metadata format for standalone scripts. Lock them with `uv lock --script`.
6. **Use `--exclude-newer` for reproducibility** — pin resolution dates to avoid unexpected updates.
7. **Use workspaces for multi-package projects** — share a single lockfile across packages.
8. **Leverage `uv pip compile` for migration** — gradually migrate from `requirements.txt` to `uv.lock` while keeping pip compatibility.
9. **Use `--no-install-project`/`--no-install-workspace` in Docker** — optimize layer caching.
10. **Run `uv cache prune --ci` in CI** — optimize cache persistence for GitHub Actions.

## Common Pitfalls

- **Project vs non-project context**: `uv run` in a directory with `pyproject.toml` will install the project. Use `--no-project` to run scripts independently.
- **`uvx` vs `uv run`**: `uvx` runs tools in isolation. Use `uv run` when the tool needs project context (e.g., `pytest`, `mypy`).
- **Inline script metadata requires `dependencies` field**: Even if empty, `dependencies = []` must be present.
- **Self-update only works with standalone install**: pip/Homebrew installs must use their respective update methods.
- **`--python-platform` affects wheel selection**: Built distributions may be incompatible with the actual platform.
- **`uv.lock` is cross-platform**: Single lockfile works for all platforms; multiple versions of a package may be present.
- **`uv pip install` does not invoke pip**: It's a native implementation; some edge cases differ.

## Version Notes

- uv is actively developed by Astral.
- Python distributions come from the Astral `python-build-standalone` project (not official Python.org binaries).
- PyPy is supported for Python management.
- The `uv.lock` format is TOML-based and human-readable but not intended for manual editing.
- `uv python upgrade` patch version upgrading is in preview.
