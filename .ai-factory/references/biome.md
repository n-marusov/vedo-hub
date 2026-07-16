# Biome Reference

> Source: https://biomejs.dev/
> Created: 2026-07-16
> Updated: 2026-07-16

## Overview

Biome is a high-performance toolchain for web projects, written in Rust, designed as a unified replacement for ESLint, Prettier, and more. It provides formatting, linting, and code analysis for JavaScript, TypeScript, JSX, TSX, JSON, HTML, CSS, and GraphQL. Biome scores 97% compatibility with Prettier while being ~35x faster, and features over 500 lint rules from ESLint, TypeScript ESLint, and other sources. It is built with an architecture inspired by rust-analyzer, requires zero configuration to get started, and supports monorepo-friendly hierarchical configuration.

## Installation

Biome is installed as a development dependency via npm-compatible package managers:

| Package Manager | Command |
|----------------|---------|
| npm | `npm i -D -E @biomejs/biome` |
| pnpm | `pnpm add -D -E @biomejs/biome` |
| bun | `bun add -D -E @biomejs/biome` |
| deno | `deno add -D npm:@biomejs/biome` |
| yarn | `yarn add -D -E @biomejs/biome` |

> **Version pinning**: The `-E` flag pins the exact version. Biome requires pinned versions for stability.

Initialize configuration: `npx @biomejs/biome init` (creates `biome.json`).

## CLI Commands

### `biome check` — Format, lint, and organize imports

```
biome check [OPTIONS] [PATH]...
```

Combined command that runs formatter, linter, and import sorting.

Key options:
- `--write` / `--fix` — Apply safe fixes, formatting, and import sorting
- `--unsafe` — Apply unsafe fixes (use with `--write`)
- `--staged` — Only check staged files
- `--changed` — Only check changed files (vs `defaultBranch`)
- `--since <REF>` — Set base branch for `--changed`
- `--only <RULE|GROUP>` — Run only specific rules/groups
- `--skip <RULE|GROUP>` — Skip specific rules/groups
- `--watch` — Watch mode (re-run on file changes)
- `--assist-enabled` — Enable/disable assist actions
- `--format-with-errors` — Allow formatting with syntax errors
- `--profile-rules` — Profile rule execution timing
- `--stdin-file-path <PATH>` — Format code from stdin
- `--reporter <FORMAT>` — Output format (default, json, json-pretty, github, junit, summary, gitlab, checkstyle, rdjson, sarif, concise)
- `--error-on-warnings` — Exit with error code on warnings

### `biome lint` — Lint only

```
biome lint [OPTIONS] [PATH]...
```

Run only lint checks. Same options as `check` for rule selection and output formatting.

Additional options:
- `--suppress` — Fix violations with suppression comments instead of code fixes
- `--reason <STRING>` — Explanation for `--suppress`

### `biome format` — Format only

```
biome format [OPTIONS] [PATH]...
```

Run only the formatter. Supports all formatting configuration options as CLI flags.

### `biome ci` — CI mode (read-only)

```
biome ci [OPTIONS] [PATH]...
```

Read-only command for CI environments. Runs formatter, linter, and import sorting without modifying files. Supports `--changed` for differential checking.

### `biome init` — Initialize configuration

```
biome init [--jsonc]
```

Creates `biome.json` with default configuration. Use `--jsonc` to create `biome.jsonc`.

### `biome migrate` — Migrate configuration

```
biome migrate [--write]
```

Updates configuration for breaking changes. Sub-commands:
- `biome migrate prettier` — Maps Prettier config to Biome
- `biome migrate eslint` — Maps ESLint config to Biome (`--include-inspired`, `--include-nursery`)

### Other commands

| Command | Description |
|---------|-------------|
| `biome version` | Show version information |
| `biome upgrade` | Upgrade Biome to latest version (standalone/Homebrew only) |
| `biome rage` | Print debugging information (`--daemon-logs`, `--formatter`, `--linter`) |
| `biome start` | Start the Biome daemon server |
| `biome stop` | Stop the Biome daemon server |
| `biome lsp-proxy` | Language Server Protocol server |
| `biome search <PATTERN>` | EXPERIMENTAL: Search with GritQL patterns |
| `biome explain <NAME>` | Show documentation for a rule or feature |
| `biome clean` | Clean daemon logs |

### Global CLI Options

| Option | Description |
|--------|-------------|
| `--colors <off\|force>` | Set markup formatting mode |
| `--use-server` | Connect to running Biome daemon |
| `--verbose` | Print additional diagnostics |
| `--config-path <PATH>` | Path to config file or directory |
| `--max-diagnostics <NUMBER\|none>` | Cap displayed diagnostics (default: 20) |
| `--skip-parse-errors` | Skip files with syntax errors |
| `--no-errors-on-unmatched` | Silence errors when no files processed |
| `--error-on-warnings` | Exit with error code on warnings |
| `--reporter <FORMAT>` | Output format |
| `--diagnostic-level <info\|warn\|error>` | Minimum diagnostic level |
| `--log-level <none\|debug\|info\|warn\|error>` | Logging level |
| `--log-kind <pretty\|compact\|json>` | Log format |

## Configuration (`biome.json` / `biome.jsonc`)

Biome uses a JSON configuration file. Top-level structure:

```json
{
  "$schema": "./node_modules/@biomejs/biome/configuration_schema.json",
  "extends": [],
  "root": true,
  "plugins": [],
  "files": {},
  "vcs": {},
  "linter": {},
  "assist": {},
  "formatter": {},
  "javascript": {},
  "json": {},
  "css": {},
  "graphql": {},
  "grit": {},
  "html": {},
  "overrides": []
}
```

### Top-level

| Setting | Default | Description |
|---------|---------|-------------|
| `$schema` | — | Path to JSON schema file |
| `extends` | `[]` | Paths to other Biome config files to merge |
| `root` | `true` | Whether this config is a root (must set `false` for nested configs) |
| `plugins` | `[]` | GritQL plugin paths with optional file filters |

### `files`

| Setting | Default | Description |
|---------|---------|-------------|
| `includes` | `["**"]` | Glob patterns for files to process |
| `ignoreUnknown` | `false` | Don't emit diagnostics for unknown files |
| `maxSize` | `1048576` (1MB) | Maximum file size in bytes |

### `vcs` (VCS integration)

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `false` | Enable VCS integration |
| `clientKind` | `"git"` | VCS client type |
| `useIgnoreFile` | `false` | Respect VCS ignore files (.gitignore, .ignore) |
| `root` | — | Path to VCS root (relative to config) |
| `defaultBranch` | — | Main branch for `--changed` |

### `linter`

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `true` | Enable the linter |
| `includes` | — | Glob patterns for files to lint |
| `rules.preset` | `"recommended"` | Preset: `"recommended"`, `"all"`, `"none"` |

#### Rule Groups

| Group | Description |
|-------|-------------|
| `a11y` | Accessibility rules |
| `complexity` | Code complexity |
| `correctness` | Code correctness |
| `nursery` | Unstable/experimental rules (require opt-in on stable) |
| `performance` | Performance optimization |
| `security` | Security flaws |
| `style` | Consistent/idiomatic code style |
| `suspicious` | Likely incorrect or useless code |

Rule configuration example:

```json
{
  "linter": {
    "rules": {
      "correctness": {
        "noUnusedVariables": "error"
      },
      "style": {
        "useConst": "warn"
      },
      "nursery": "off",
      "suspicious": "info"
    }
  }
}
```

Severity values: `"on"` (default severity), `"off"`, `"info"`, `"warn"`, `"error"`.

### `assist`

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `true` | Enable assist actions |
| `includes` | — | Glob patterns for assist |
| `actions.source.recommended` | — | Recommended source actions (safe actions on save) |

### `formatter`

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `true` | Enable the formatter |
| `includes` | — | Glob patterns for files to format |
| `formatWithErrors` | `false` | Allow formatting files with syntax errors |
| `indentStyle` | `"tab"` | Indentation style (`"tab"` or `"space"`) |
| `indentWidth` | `2` | Indentation width (ignored for tabs) |
| `lineEnding` | `"lf"` | Line ending (`"lf"`, `"crlf"`, `"cr"`) |
| `lineWidth` | `80` | Max characters per line |
| `attributePosition` | `"auto"` | Attribute position in HTML-ish languages |
| `bracketSpacing` | `true` | Spaces between brackets and values |
| `delimiterSpacing` | `false` | Spaces inside delimiters |
| `expand` | `"auto"` | Array/object expansion (`"auto"`, `"always"`, `"never"`) |
| `trailingNewline` | `true` | Add trailing newline at end of file |
| `useEditorconfig` | `false` | Respect `.editorconfig` files |

### Language-specific settings

#### `javascript`

| Setting | Default | Description |
|---------|---------|-------------|
| `parser.unsafeParameterDecoratorsEnabled` | `false` | Support experimental parameter decorators |
| `parser.jsxEverywhere` | `true` | Allow JSX in `.js` files |
| `formatter.quoteStyle` | `"double"` | Quote style (`"single"` or `"double"`) |
| `formatter.jsxQuoteStyle` | `"double"` | JSX quote style |
| `formatter.quoteProperties` | `"asNeeded"` | Property quoting (`"asNeeded"` or `"preserve"`) |
| `formatter.trailingCommas` | `"all"` | Trailing commas (`"all"`, `"es5"`, `"none"`) |
| `formatter.semicolons` | `"always"` | Semicolons (`"always"` or `"asNeeded"`) |
| `formatter.arrowParentheses` | `"always"` | Arrow function parentheses |
| `formatter.bracketSameLine` | `false` | Closing `>` on last attribute line |
| `formatter.bracketSpacing` | `true` | Spaces between brackets |
| `formatter.delimiterSpacing` | `false` | Spaces inside delimiters |
| `formatter.attributePosition` | `"auto"` | Jsx element single attribute per line |
| `formatter.expand` | `"auto"` | Expand nested data structures on multiple lines |
| `formatter.operatorLinebreak` | `"after"` | Bin expression line break placement |
| `formatter.trailingNewline` | `true` | Trailing newline |
| `formatter.indentStyle` | `"tab"` | Indentation style |
| `formatter.indentWidth` | `2` | Indentation width |
| `formatter.lineEnding` | `"lf"` | Line ending |
| `formatter.lineWidth` | `80` | Line width |
| `formatter.enabled` | `true` | Enable JS formatter |
| `linter.enabled` | `true` | Enable JS linter |
| `assist.enabled` | `true` | Enable JS assist |
| `globals` | `[]` | Global names to ignore |
| `jsxRuntime` | `"transparent"` | JSX runtime (`"transparent"` or `"reactClassic"`) |
| `resolver.experimentalPnpmCatalogs` | `false` | Resolve pnpm catalogs |
| `experimentalEmbeddedSnippetsEnabled` | `false` | Format embedded snippets |

#### `json`

| Setting | Default | Description |
|---------|---------|-------------|
| `parser.allowComments` | `false` | Allow comments in `.json` |
| `parser.allowTrailingCommas` | `false` | Allow trailing commas in `.json` |
| `formatter.enabled` | `true` | Enable JSON formatter |
| `formatter.trailingCommas` | `"none"` | Trailing commas (`"none"` or `"all"`) |
| `formatter.bracketSpacing` | `true` | Spaces between brackets |
| `formatter.delimiterSpacing` | `false` | Spaces inside brackets |
| `formatter.expand` | `"auto"` | Expand arrays/objects |
| `formatter.trailingNewline` | `true` | Trailing newline |
| `formatter.indentStyle` | `"tab"` | Indentation style |
| `formatter.indentWidth` | `2` | Indentation width |
| `formatter.lineEnding` | `"lf"` | Line ending |
| `formatter.lineWidth` | `80` | Line width |
| `linter.enabled` | `true` | Enable JSON linter |
| `assist.enabled` | `true` | Enable JSON assist |

#### `css`

| Setting | Default | Description |
|---------|---------|-------------|
| `parser.cssModules` | `false` | Parse CSS Modules features |
| `parser.tailwindDirectives` | `false` | Parse Tailwind CSS 4.0 directives |
| `formatter.enabled` | `false` | Enable CSS formatter |
| `formatter.quoteStyle` | `"double"` | Quote style |
| `formatter.delimiterSpacing` | `false` | Spaces inside parentheses/brackets |
| `formatter.trailingNewline` | `true` | Trailing newline |
| `formatter.indentStyle` | `"tab"` | Indentation style |
| `formatter.indentWidth` | `2` | Indentation width |
| `formatter.lineEnding` | `"lf"` | Line ending |
| `formatter.lineWidth` | `80` | Line width |
| `linter.enabled` | `true` | Enable CSS linter |
| `assist.enabled` | `true` | Enable CSS assist |

#### `graphql`

| Setting | Default | Description |
|---------|---------|-------------|
| `formatter.enabled` | `false` | Enable GraphQL formatter |
| `formatter.quoteStyle` | `"double"` | Quote style |
| `formatter.trailingNewline` | `true` | Trailing newline |
| `formatter.indentStyle` | `"tab"` | Indentation style |
| `formatter.indentWidth` | `2` | Indentation width |
| `formatter.lineEnding` | `"lf"` | Line ending |
| `formatter.lineWidth` | `80` | Line width |
| `linter.enabled` | `true` | Enable GraphQL linter |
| `assist.enabled` | `true` | Enable GraphQL assist |

#### `html`

| Setting | Default | Description |
|---------|---------|-------------|
| `experimentalFullSupportEnabled` | — | Full support for Vue/Svelte/Astro files |
| `parser.interpolation` | `false` | Parse `{{ expression }}` in `.html` |
| `parser.vue` | `false` | Parse Vue-specific syntax |
| `formatter.enabled` | `false` | Enable HTML formatter (experimental) |
| `formatter.attributePosition` | `"auto"` | Attribute position |
| `formatter.bracketSameLine` | `false` | Closing `>` on last line |
| `formatter.whitespaceSensitivity` | `"css"` | Whitespace handling |
| `formatter.indentScriptAndStyle` | `false` | Indent `<script>` and `<style>` content |
| `formatter.selfCloseVoidElements` | `"never"` | Self-closing void elements |
| `formatter.trailingNewline` | `true` | Trailing newline |
| `linter.enabled` | `true` | Enable HTML linter |
| `assist.enabled` | `true` | Enable HTML assist |

### `overrides`

Override settings per file pattern:

```json
{
  "overrides": [
    {
      "includes": ["generated/**"],
      "formatter": {
        "lineWidth": 160,
        "indentStyle": "space"
      }
    },
    {
      "includes": ["shims/**"],
      "linter": {
        "enabled": false
      }
    }
  ]
}
```

## Key Environment Variables

| Variable | Purpose |
|----------|---------|
| `BIOME_CONFIG_PATH` | Path to configuration file |
| `BIOME_LOG_FILE` | Log file path |
| `BIOME_LOG_PREFIX_NAME` | Log prefix name (default: server.log) |
| `BIOME_LOG_PATH` | Daemon log directory |
| `BIOME_LOG_LEVEL` | Log level (none, debug, info, warn, error) |
| `BIOME_LOG_KIND` | Log format (pretty, compact, json) |
| `BIOME_WATCHER_KIND` | File watcher type (polling, recommended, none) |
| `BIOME_WATCHER_POLLING_INTERVAL` | Polling interval in ms |
| `BIOME_THREADS` | Number of threads for CI |

## Supported Languages

| Language | Formatter | Linter | Notes |
|----------|-----------|--------|-------|
| JavaScript | ✅ | ✅ | Includes JSX |
| TypeScript | ✅ | ✅ | Full support |
| JSX/TSX | ✅ | ✅ | |
| JSON | ✅ | ✅ | Comments & trailing commas opt-in |
| CSS | ✅ | ✅ | Formatter opt-in (enabled: false) |
| HTML | ✅ | ✅ | Formatter experimental (enabled: false) |
| GraphQL | ✅ | ✅ | Formatter opt-in (enabled: false) |
| GritQL | ✅ | ✅ | Pattern files |

## Best Practices

1. **Pin Biome version** — use `-E` flag during installation to prevent unexpected behavior changes.
2. **Use `biome ci` in CI** — read-only check that fails on unformatted code or lint violations.
3. **Use `--changed` for differential checks** — combine with `--since` or `vcs.defaultBranch` for faster CI.
4. **Use `migrate` when switching from Prettier/ESLint** — run `npx @biomejs/biome migrate prettier` and `npx @biomejs/biome migrate eslint`.
5. **Use `overrides` for generated files** — disable linting or increase line width for generated code.
6. **Start with recommended preset** and selectively enable individual rules/groups as needed.
7. **Enable VCS integration** (`vcs.enabled`, `vcs.useIgnoreFile`) to respect `.gitignore`.
8. **Enable CSS/HTML/GraphQL formatter only when needed** — they are opt-in for stability reasons.
9. **Use `biome check --write`** for a single command that formats, lints, and organizes imports.
10. **Use `--reporter` for machine-readable output** — `json`, `github`, `gitlab`, `sarif` for integration.

## Common Pitfalls

- **`npx @biomejs/biome check --write` must be used instead of separate format + lint** — unlike ESLint/Prettier, Biome combines both in the `check` command.
- **Biome does not support custom ESLint plugins** — all rules are natively implemented in Rust.
- **`biome.json` root defaults to `true`** — nested configs must explicitly set `"root": false`.
- **`formatter.expand` defaults to `"auto"`** — but `package.json` always uses `"always"`.
- **`indentStyle` defaults to `"tab"`** — unlike Prettier which defaults to spaces.
- **`lineWidth` defaults to `80`** — not 88 like Prettier/Ruff.
- **`css.formatter.enabled` and `graphql.formatter.enabled` default to `false`** — must be explicitly enabled.
- **HTML formatter is experimental** — breaking changes may occur.
- **`biome upgrade` only works for standalone/Homebrew installs** — npm installs must use package manager upgrade.
- **`--unsafe` fixes may change runtime behavior** — review before applying.

## Version Notes

- Biome is in active development by the Biome community (formerly Rome).
- The project follows semantic versioning.
- HTML formatter is experimental and disabled by default.
- CSS, GraphQL, and Grit support is available but some features are opt-in.
- Nursery rules are unstable and require explicit opt-in on stable versions.
- The v2.x version line introduced major improvements over v1.x.
- `linter.rules.recommended` is deprecated in favor of `linter.rules.preset`.
