# pnpm v11 Settings Reference

> Source: https://pnpm.io/settings, https://pnpm.io/package_json
> Created: 2026-07-29
> Updated: 2026-07-29

## Overview

pnpm v11 (current version: 11.17.0) introduced a **breaking change**: the `pnpm` field in `package.json` is **no longer read**. Settings that were previously configured under `pnpm.{key}` in `package.json` must be moved to either `pnpm-workspace.yaml` or `.npmrc`.

The deprecation warning appears on `pnpm install`:
```
[WARN] The "pnpm" field in package.json is no longer read by pnpm.
The following keys were ignored: "pnpm.overrides".
See https://pnpm.io/settings for the new home of each setting.
```

## Settings Migration (package.json → new location)

| Old `package.json` key | New location | New format |
|------------------------|-------------|------------|
| `pnpm.overrides` | `pnpm-workspace.yaml` | `overrides:` block |
| `pnpm.packageExtensions` | `.npmrc` or `pnpm-workspace.yaml` | `package-extensions.` prefix |
| `pnpm.peerDependencyRules` | `.npmrc` | `peer-dependency-rules.*` keys |
| `pnpm.onlyBuiltDependencies` | `.npmrc` | `only-built-dependencies=` |
| `pnpm.allowedBuiltinDependencies` | `.npmrc` | `allowed-builtin-dependencies=` |
| `pnpm.executionEnv` | `.npmrc` | N/A — removed? |

## Overrides in `pnpm-workspace.yaml`

**This is the only supported way to set overrides in pnpm v11.**

### For workspace projects (monorepos)

Create or update `pnpm-workspace.yaml` at the project root:

```yaml
packages:
  - "apps/*"
  - "packages/*"

overrides:
  brace-expansion: ^5.0.8
  postcss: ^8.5.18
```

### For single-package projects

`pnpm-workspace.yaml` works for single packages too:

```yaml
overrides:
  brace-expansion: ^5.0.8
  postcss: ^8.5.18
```

⚠️ **Important:** After adding/updating overrides, delete the lockfile (`pnpm-lock.yaml`) and `node_modules/.pnpm/`, then run `pnpm install` to force full re-resolution. Otherwise pnpm may skip the resolution step and reuse the old lockfile.

## Build Script Approval (`onlyBuiltDependencies`)

pnpm v11 **blocks all build scripts** by default (security hardening). Packages that need to run `postinstall` or other build scripts must be explicitly approved.

Error when blocked:
```
[ERR_PNPM_IGNORED_BUILDS] Ignored build scripts: @biomejs/biome@1.9.4, esbuild@0.25.12
Run "pnpm approve-builds" to pick which dependencies should be allowed to run scripts.
```

### Adding to pnpm-workspace.yaml

**This is the only way to configure it in pnpm v11.** `.npmrc` with `only-built-dependencies=...` is not recognized.

```yaml
overrides:
  brace-expansion: ^5.0.8

onlyBuiltDependencies:
  - '@biomejs/biome'
  - esbuild
  - lefthook
  - vue-demi
```

### If build approval state is corrupted

If packages were previously installed with `--ignore-scripts` and pnpm v11 cached the "ignored" state, run:

```sh
pnpm store prune  # clears cached packages
rm -rf pnpm-lock.yaml node_modules/.pnmpnpm install
```

### For CI with no build scripts

CI/Docker builds should use `--ignore-scripts` flag (no config change needed):

```sh
pnpm install --ignore-scripts
```

### Verification

Check the lockfile that the override was applied:

```yaml
minimatch@9.0.9:
  dependencies:
    brace-expansion: 5.0.8  # ← expected: 5.0.8, not 2.1.2
```

Run audit to confirm:

```sh
pnpm audit --audit-level=high
# → "No known vulnerabilities found"
```

## Settings NOT affected

The following standard npm `package.json` fields remain unchanged:

- `name`, `version`, `private`, `type` — standard npm fields
- `scripts` — unchanged
- `dependencies`, `devDependencies` — unchanged
- `packageManager` — still respected by corepack (set to `pnpm@11.x.x`)
- `overrides` — the npm-standard `"overrides"` at the root of `package.json` is deprecated in pnpm v11. Use `pnpm-workspace.yaml` instead.

## Key Differences from pnpm v10

| Feature | pnpm v10 | pnpm v11 |
|---------|----------|----------|
| `pnpm.overrides` in package.json | ✅ | ❌ deprecated |
| `"overrides"` at root of package.json | ✅ (pnpm alias) | ❌ deprecated |
| `pnpm-workspace.yaml overrides:` | ❌ not supported | ✅ required |
| `.npmrc` `pnpm.overrides.*` | ✅ | ❌ not effective |

## Best Practices

1. **Use `pnpm-workspace.yaml`** for all pnpm-specific settings in v11
2. **Set `packageManager`** in `package.json` for corepack version pinning:
   ```json
   "packageManager": "pnpm@11.17.0"
   ```
3. **Clean rebuild after config changes**: always delete `pnpm-lock.yaml` + `node_modules/.pnpm/` when changing overrides
4. **Check `.npmrc`** for leftover `pnpm.overrides.*` config that will be ignored
5. **Keep `overrides` in only one place** (pnpm-workspace.yaml) to avoid confusion

## Common Pitfalls

- **pnpm.overrides in .npmrc silently ignored**: No warning is emitted for `.npmrc` config that uses `pnpm.overrides.*` — it's simply not applied
- **Lockfile not regenerated**: `pnpm install` skips resolution if it detects the existing lockfile is "up to date". Always delete the lockfile when changing overrides
- **Multiple override locations**: Having `overrides` in both `package.json` and `pnpm-workspace.yaml` leads to confusion — use only `pnpm-workspace.yaml`
