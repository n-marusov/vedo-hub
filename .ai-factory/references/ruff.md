# Ruff Reference

> Source: https://docs.astral.sh/ruff/
> Created: 2026-07-16
> Updated: 2026-07-16

## Overview

Ruff is an extremely fast Python linter and code formatter, written in Rust, backed by Astral (the creators of uv). It is 10-100x faster than existing linters (like Flake8) and formatters (like Black). Ruff supports over 900 built-in lint rules with native re-implementations of popular Flake8 plugins, and can replace Flake8 (plus dozens of plugins), Black, isort, pydocstyle, pyupgrade, autoflake, and more. It features built-in caching, automatic fix support, pyproject.toml configuration, Python 3.14 compatibility, first-party editor integrations (VS Code et al.), and monorepo-friendly hierarchical configuration.

## Installation

| Method | Command |
|--------|---------|
| uvx (ephemeral) | `uvx ruff check` or `uvx ruff format` |
| uv tool | `uv tool install ruff@latest` |
| uv (project) | `uv add --dev ruff` |
| pip | `pip install ruff` |
| pipx | `pipx install ruff` |
| Standalone (macOS/Linux) | `curl -LsSf https://astral.sh/ruff/install.sh \| sh` |
| Standalone (Windows) | `powershell -c "irm https://astral.sh/ruff/install.ps1 \| iex"` |
| Homebrew | `brew install ruff` |
| Conda | `conda install -c conda-forge ruff` |
| Docker | `ghcr.io/astral-sh/ruff` |
| Alpine | `apk add ruff` |
| Arch Linux | `pacman -S ruff` |
| pkgx | `pkgx install ruff` |

## Commands

### `ruff check` — Lint Python files

```
ruff check [OPTIONS] [FILES]...
```

Lints all discovered Python files. Key options:

- `--fix` — Apply safe fixes
- `--unsafe-fixes` — Include unsafe fixes
- `--fix-only` — Apply fixes without reporting remaining violations
- `--diff` — Show diff without modifying files
- `--watch` / `-w` — Watch mode (re-lint on file changes)
- `--select <RULE_CODE>` — Comma-separated rule codes to enable
- `--ignore <RULE_CODE>` — Comma-separated rule codes to disable
- `--extend-select <RULE_CODE>` — Add rules on top of current selection
- `--per-file-ignores` — File-specific ignore patterns
- `--output-format` — Output format (concise, full, json, json-lines, junit, grouped, github, gitlab, pylint, rdjson, azure, sarif)
- `--preview` — Enable preview rules and fixes
- `--show-files` — Show files that would be checked
- `--show-settings` — Show resolved settings
- `--add-noqa` — Auto-add noqa directives
- `--exit-zero` — Exit with code 0 even if violations found
- `--exit-non-zero-on-fix` — Exit non-zero if files were modified
- `--target-version` — Minimum Python version
- `--statistics` — Show violation counts per rule
- `--ignore-noqa` — Ignore noqa comments

Exit codes: 0 = no violations, 1 = violations found, 2 = abnormal termination.

### `ruff format` — Format Python code

```
ruff format [OPTIONS] [FILES]...
```

Formats Python files using the Ruff formatter (drop-in replacement for Black). Key options:

- `--check` — Check without writing (exit 1 if unformatted)
- `--diff` — Show diff without writing
- `--preview` — Enable unstable formatting
- `--target-version` — Minimum Python version
- `--line-length` — Override line length
- `--range` — Format a specific range (<start_line>:<start_col>-<end_line>:<end_col>)
- `--exit-non-zero-on-format` — Exit non-zero if files were modified

### `ruff rule` — Explain rules

```
ruff rule [RULE_CODE]
```

Display documentation for a specific rule or all rules.

### `ruff config` — List configuration options

```
ruff config [OPTIONS]
```

List or describe available configuration options.

### `ruff linter` — List upstream linters

```
ruff linter
```

List all supported upstream linters.

### `ruff clean` — Clear caches

```
ruff clean [OPTIONS]
```

Clear any caches in the current directory and subdirectories.

### `ruff server` — Language server

```
ruff server [OPTIONS]
```

Run the language server (LSP).

### `ruff analyze` — Dependency analysis

```
ruff analyze [OPTIONS]
```

Run analysis over Python source code (dependency graph).

### `ruff version` — Display version

```
ruff version
```

### `ruff generate-shell-completion` — Shell completions

```
ruff generate-shell-completion <SHELL>
```

Supports bash, elvish, fig, fish, powershell, zsh.

## Configuration

Ruff can be configured through `pyproject.toml`, `ruff.toml`, or `.ruff.toml` files. All three implement an equivalent schema (in `ruff.toml`/`.ruff.toml`, the `[tool.ruff]` header is omitted).

### Default configuration behavior

```toml
[tool.ruff]
exclude = [".bzr", ".direnv", ".eggs", ".git", ".git-rewrite", ".hg",
    ".ipynb_checkpoints", ".mypy_cache", ".nox", ".pants.d", ".pyenv",
    ".pytest_cache", ".pytype", ".ruff_cache", ".svn", ".tox", ".venv",
    ".vscode", "__pypackages__", "_build", "buck-out", "build", "dist",
    "node_modules", "site-packages", "venv"]
line-length = 88
indent-width = 4
target-version = "py310"

[tool.ruff.lint]
select = ["E4", "E7", "E9", "F"]
ignore = []
fixable = ["ALL"]
unfixable = []
dummy-variable-rgx = "^(_+|(_+[a-zA-Z0-9_]*[a-zA-Z0-9]+?))$"

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
skip-magic-trailing-comma = false
line-ending = "auto"
docstring-code-format = false
docstring-code-line-length = "dynamic"
```

### Key top-level settings

| Setting | Default | Description |
|---------|---------|-------------|
| `exclude` | [see above] | File patterns to exclude |
| `extend-exclude` | `[]` | Additional exclusion patterns |
| `include` | `["*.py", "*.pyi", "*.pyw", "*.ipynb", "*.md", "**/pyproject.toml"]` | File patterns to include |
| `extend-include` | `[]` | Additional inclusion patterns |
| `line-length` | `88` | Line length for formatting and E501 |
| `indent-width` | `4` | Spaces per indentation level |
| `target-version` | `"py310"` | Minimum Python version |
| `preview` | `false` | Enable preview rules and formatting |
| `src` | `[".", "src"]` | Directories for first-party import resolution |
| `respect-gitignore` | `true` | Respect .gitignore exclusions |
| `fix` | `false` | Enable fix by default |
| `unsafe-fixes` | `null` | Enable unsafe fixes |
| `force-exclude` | `false` | Enforce exclude even for explicit paths |
| `namespace-packages` | `[]` | Directories treated as namespace packages |
| `builtins` | `[]` | Additional builtin references |
| `cache-dir` | `".ruff_cache"` | Cache directory path |
| `output-format` | `"full"` | Output format |
| `show-fixes` | `false` | Show enumeration of fixed violations |
| `required-version` | `null` | Enforce Ruff version requirement |
| `extension` | `{}` | Custom file extension mappings |
| `per-file-target-version` | `{}` | Per-file Python version overrides |

### Lint settings (under `[tool.ruff.lint]`)

| Setting | Default | Description |
|---------|---------|-------------|
| `select` | `["E4", "E7", "E9", "F"]` | Enabled rule codes |
| `extend-select` | `[]` | Additional rules to enable |
| `ignore` | `[]` | Disabled rule codes |
| `fixable` | `["ALL"]` | Rules eligible for fix |
| `unfixable` | `[]` | Rules ineligible for fix |
| `extend-fixable` | `[]` | Additional fixable rules |
| `per-file-ignores` | `{}` | File-specific ignore patterns |
| `extend-per-file-ignores` | `{}` | Additional per-file ignores |
| `dummy-variable-rgx` | `"^(_+\|...)$"` | Regex for dummy variable names |
| `external` | `[]` | External rule codes to preserve in noqa |
| `allowed-confusables` | `[]` | Allowed ambiguous Unicode characters |
| `typing-modules` | `[]` | Modules treated as typing re-exports |
| `typing-extensions` | `true` | Allow typing_extensions fallback |
| `logger-objects` | `[]` | Objects treated as loggers |
| `task-tags` | `["TODO", "FIXME", "XXX"]` | Task tags for comment detection |
| `future-annotations` | `false` | Allow adding `from __future__ import annotations` |
| `explicit-preview-rules` | `false` | Require exact codes for preview rules |
| `extend-safe-fixes` | `[]` | Promote unsafe fixes to safe |
| `extend-unsafe-fixes` | `[]` | Demote safe fixes to unsafe |

### Format settings (under `[tool.ruff.format]`)

| Setting | Default | Description |
|---------|---------|-------------|
| `quote-style` | `"double"` | Quote style (double, single, preserve) |
| `indent-style` | `"space"` | Indentation style (space, tab) |
| `skip-magic-trailing-comma` | `false` | Ignore magic trailing commas |
| `line-ending` | `"auto"` | Line ending (auto, lf, cr-lf, native) |
| `docstring-code-format` | `false` | Format code in docstrings |
| `docstring-code-line-length` | `"dynamic"` | Line length for docstring code |
| `preview` | `false` | Enable preview style formatting |
| `nested-string-quote-style` | `"alternating"` | Quote style for nested f-strings |
| `exclude` | `[]` | File patterns to exclude from formatting |

### Plugin-specific settings

Sub-sections under `[tool.ruff.lint]` for plugin configuration:

- `flake8-annotations` — `suppress-none-returning`, `suppress-dummy-args`, `allow-star-arg-any`, `mypy-init-return`, `ignore-fully-untyped`
- `flake8-bandit` — `check-typed-exception`, `hardcoded-tmp-directory`, `allowed-markup-calls`, `extend-markup-names`
- `flake8-boolean-trap` — `extend-allowed-calls`
- `flake8-bugbear` — `extend-immutable-calls`
- `flake8-builtins` — `ignorelist`, `allowed-modules`, `strict-checking`
- `flake8-comprehensions` — `allow-dict-calls-with-keyword-arguments`
- `flake8-copyright` — `author`, `min-file-size`, `notice-rgx`
- `flake8-errmsg` — `max-string-length`
- `flake8-gettext` — `function-names`, `extend-function-names`
- `flake8-implicit-str-concat` — `allow-multiline`
- `flake8-import-conventions` — `aliases`, `extend-aliases`, `banned-aliases`, `banned-from`
- `flake8-pytest-style` — `fixture-parentheses`, `mark-parentheses`, `parametrize-names-type`, `parametrize-values-type`, `parametrize-values-row-type`, `raises-require-match-for`, `warns-require-match-for`
- `flake8-quotes` — `inline-quotes`, `multiline-quotes`, `docstring-quotes`, `avoid-escape`
- `flake8-self` — `ignore-names`, `extend-ignore-names`
- `flake8-tidy-imports` — `ban-relative-imports`, `banned-api`, `banned-module-level-imports`, `ban-lazy`, `require-lazy`
- `flake8-type-checking` — `exempt-modules`, `quote-annotations`, `runtime-evaluated-base-classes`, `runtime-evaluated-decorators`, `strict`
- `flake8-unused-arguments` — `ignore-variadic-names`
- `isort` — `force-single-line`, `force-sort-within-sections`, `force-wrap-aliases`, `combine-as-imports`, `order-by-type`, `case-sensitive`, `section-order`, `default-section`, `known-first-party`, `known-third-party`, `known-local-folder`, `extra-standard-library`, `src`, `lines-after-imports`, `lines-between-types`, `split-on-trailing-comma`, `required-imports`, `force-to-top`, `no-lines-before`, `no-sections`, `from-first`, `length-sort`, `relative-imports-order`, `single-line-exclusions`, `sections`, `classes`, `constants`, `variables`, `import-heading`, `detect-same-package`
- `mccabe` — `max-complexity` (default: 10)
- `pep8-naming` — `ignore-names`, `extend-ignore-names`, `classmethod-decorators`, `staticmethod-decorators`
- `pycodestyle` — `max-doc-length`, `max-line-length`, `ignore-overlong-task-comments`
- `pydoclint` — `ignore-one-line-docstrings`
- `pydocstyle` — `convention` (google, numpy, pep257), `ignore-decorators`, `property-decorators`, `ignore-var-parameters`
- `pyflakes` — `allowed-unused-imports`, `extend-generics`
- `pylint` — `max-args` (5), `max-positional-args` (5), `max-returns` (6), `max-branches` (12), `max-statements` (50), `max-locals` (15), `max-nested-blocks` (5), `max-bool-expr` (5), `max-public-methods` (20), `max-statements-in-try` (5), `allow-magic-value-types`, `allow-dunder-method-names`
- `pyupgrade` — `keep-runtime-typing`
- `ruff` — `parenthesize-tuple-in-subscript`, `strictly-empty-init-modules`

## Rule Categories

Ruff supports over 900 lint rules organized by prefix:

| Prefix | Source | Description |
|--------|--------|-------------|
| `F` | Pyflakes | Core logic errors, undefined names, unused imports |
| `E` / `W` | pycodestyle | PEP 8 style conventions (errors / warnings) |
| `D` | pydocstyle | Docstring conventions |
| `I` | isort | Import sorting |
| `N` | pep8-naming | Naming conventions |
| `UP` | pyupgrade | Modern Python syntax upgrades |
| `B` | flake8-bugbear | Common bugs and design issues |
| `SIM` | flake8-simplify | Code simplification |
| `C4` | flake8-comprehensions | Comprehensions and generators |
| `PL` | Pylint | Pylint rules (C/R/E/W) |
| `RUF` | Ruff-specific | Custom Ruff rules |
| `S` | flake8-bandit | Security issues |
| `A` | flake8-builtins | Builtin shadowing |
| `ANN` | flake8-annotations | Type annotation requirements |
| `ARG` | flake8-unused-arguments | Unused arguments |
| `ASYNC` | flake8-async | Async/await best practices |
| `BLE` | flake8-blind-except | Blind except clauses |
| `C90` | mccabe | McCabe complexity |
| `COM` | flake8-commas | Comma style |
| `CPY` | flake8-copyright | Copyright notices |
| `DJ` | flake8-django | Django best practices |
| `DOC` | pydoclint | Docstring parameter documentation |
| `DTZ` | flake8-datetimez | Timezone-aware datetime |
| `EM` | flake8-errmsg | Exception messages |
| `ERA` | eradicate | Commented-out code |
| `EXE` | flake8-executable | Shebang and executable bits |
| `FA` | flake8-future-annotations | `from __future__ import annotations` |
| `FAST` | FastAPI | FastAPI-specific rules |
| `FBT` | flake8-boolean-trap | Boolean trap arguments |
| `FIX` | flake8-fixme | FIXME/TODO/XXX/HACK |
| `FLY` | flynt | f-string conversion |
| `FURB` | refurb | Code quality improvements |
| `G` | flake8-logging-format | Logging format strings |
| `ICN` | flake8-import-conventions | Import alias conventions |
| `INP` | flake8-no-pep420 | Implicit namespace packages |
| `INT` | flake8-gettext | Internationalization |
| `ISC` | flake8-implicit-str-concat | String concatenation |
| `LOG` | flake8-logging | Logging best practices |
| `NPY` | NumPy | NumPy-specific rules |
| `PD` | pandas-vet | pandas best practices |
| `PERF` | Perflint | Performance improvements |
| `PGH` | pygrep-hooks | pygrep hooks |
| `PIE` | flake8-pie | Pie fixes |
| `PLC`/`PLE`/`PLR`/`PLW` | Pylint | Pylint convention/error/refactor/warning |
| `PT` | flake8-pytest-style | pytest style |
| `PTH` | flake8-use-pathlib | pathlib over os.path |
| `PYI` | flake8-pyi | Stub file rules |
| `Q` | flake8-quotes | Quote style |
| `RET` | flake8-return | Return statements |
| `RSE` | flake8-raise | Raise statements |
| `SLF` | flake8-self | Private member access |
| `SLOT` | flake8-slots | `__slots__` usage |
| `T10` | flake8-debugger | Debugger statements |
| `T20` | flake8-print | Print statements |
| `TC` | flake8-type-checking | Type-checking block optimization |
| `TD` | flake8-todos | TODO formatting |
| `TID` | flake8-tidy-imports | Import tidiness |
| `TRY` | tryceratops | Exception handling |
| `YTT` | flake8-2020 | sys.version/six version checks |
| `AIR` | Airflow | Apache Airflow rules |

### Special rules

- `ALL` — Enable all rules (may include conflicting pydocstyle rules; Ruff auto-resolves conflicts)
- `🧪` — Preview (unstable)
- `⚠️` — Deprecated (will be removed)
- `❌` — Removed (documentation only)
- `🛠️` — Auto-fixable via `--fix`

## Error Suppression

### Inline noqa

```python
x = 1  # noqa: F841      # Suppress specific rule
x = 1  # noqa: E741, F841 # Suppress multiple rules
x = 1  # noqa              # Suppress all violations
```

### File-level

```python
# ruff: noqa              # Suppress all violations in file
# ruff: noqa: F841        # Suppress specific rule in file
```

### Block-level (range suppression)

```python
# ruff: disable[E501]
VALUE = "long string..."
# ruff: enable[E501]
```

### Line-level (preview)

```python
# ruff: ignore[unused-function-argument]  # Covers entire function signature
def foo(arg1, arg2): pass
```

## Fix Safety

Ruff classifies fixes as:
- **Safe** — Retain runtime behavior (enabled by default)
- **Unsafe** — May change runtime behavior or error types

Control unsafe fixes via `--unsafe-fixes` flag or `unsafe-fixes = true` config.

Override per-rule with `extend-safe-fixes` and `extend-unsafe-fixes`.

## Key Environment Variables

| Variable | Purpose |
|----------|---------|
| `RUFF_OUTPUT_FORMAT` | Output serialization format |
| `RUFF_OUTPUT_FILE` | Output file path |
| `RUFF_CACHE_DIR` | Cache directory |
| `RUFF_NO_CACHE` | Disable cache reads |
| `NO_COLOR` | Disable colored output |
| `FORCE_COLOR` | Force colored output |

## Best Practices

1. **Start with defaults** (`select = ["E4", "E7", "E9", "F"]`) and add categories gradually.
2. **Use `ruff format` alongside `ruff check`** — run `ruff check --select I --fix` then `ruff format` for import sorting + formatting.
3. **Configure `src` for correct import categorization** — especially in `src`-layout projects.
4. **Use `--add-noqa` for migration** — automatically annotate existing violations when adopting Ruff on a legacy codebase.
5. **Use `convention` for docstrings** — set `[tool.ruff.lint.pydocstyle] convention = "google"` (or numpy/pep257).
6. **Use `extend-select` to add rules** rather than replacing `select` entirely, to keep the defaults.
7. **Run `ruff check --show-settings`** to debug configuration issues.
8. **Use `per-file-ignores` for test files** — e.g., ignore `D` rules in test directories.
9. **Use `preview` mode to try new rules early** — enable `preview = true` and give feedback.
10. **Disable conflicting lint rules when using the formatter** — avoid `W191`, `E1xx`, `Q000`-`Q004`, `COM812`, `COM819`, `D203`, `D206`, `D300`.

## Common Pitfalls

- **`ruff format` is not a hard line-wrap** — formatted lines may exceed `line-length` in some cases (e.g., comments, strings).
- **`ruff check` and `ruff format` are independent** — you need to run both to lint and format.
- **Import sorting is a lint rule, not a formatter feature** — use `ruff check --select I --fix`.
- **Unsafe fixes can change exception types** — review before applying (e.g., `list(...)[0]` → `next(iter(...))` changes `IndexError` to `StopIteration`).
- **Target-version affects rules** — Ruff won't suggest features newer than `target-version`.
- **`ALL` will include new rules on upgrade** — use with caution; prefer explicit selection.
- **Jupyter Notebooks require full notebook context** — avoid `source.*` code actions; use `notebook.*` instead.
- **Some isort settings conflict with the formatter** — avoid `force-single-line`, `force-wrap-aliases`, `lines-after-imports`, `lines-between-types`, `split-on-trailing-comma` with non-default values.

## Version Notes

- Ruff is extremely actively developed by Astral.
- Supports Python 3.7–3.14 (not Python 2).
- Over 900 rules with many more in preview.
- The Ruff formatter targets Black compatibility (>99.9% identical on Black-formatted code).
- Ruff supports linting and formatting of Jupyter Notebooks (.ipynb) and Markdown files.
- Ruff does not support custom/third-party plugins (planned for the future).
- Ruff is meant to complement type checkers (Mypy, Pyright, Pyre), not replace them.
- The project follows semantic versioning, with preview features promoted to stable through minor releases.
