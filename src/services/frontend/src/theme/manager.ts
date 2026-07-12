// @hlv:artifact code-frontend implements spec-ux-theme-001
// @ctx: theme manager — input validation, mode switching, override application per UX-THEME-001
// @hlv:sec [INPUT_VALIDATION] — validates mode enum and hex color overrides before applying

import {
  ERROR_CODES,
  THEME_CONTRACT_VERSION,
  type ThemeConfig,
  ThemeError,
  type ThemeMode,
  type ThemeOverride,
  type ThemeResponse,
  VALID_THEME_MODES,
  isValidHexColor
} from './types'

// @ctx: structured logging for theme operations (observability constraint)
const log = {
  info: (_msg: string, _ctx: Record<string, unknown>) => {},
  warn: (msg: string, ctx: Record<string, unknown>) => {
    console.warn(JSON.stringify({ level: 'warn', msg, ...ctx, ts: new Date().toISOString() }))
  },
  error: (msg: string, ctx: Record<string, unknown>) => {
    console.error(JSON.stringify({ level: 'error', msg, ...ctx, ts: new Date().toISOString() }))
  }
}

/**
 * Validate theme mode input per UX-THEME-001 contract.
 * @hlv THEME_INVALID_MODE
 */
export function validateMode(mode: unknown): ThemeMode {
  if (typeof mode !== 'string' || !VALID_THEME_MODES.includes(mode as ThemeMode)) {
    const error = new ThemeError(
      ERROR_CODES.THEME_INVALID_MODE,
      `Unsupported theme mode: ${String(mode)}. Use 'light' or 'dark'.`
    )
    log.error('theme.mode.invalid', { mode, code: error.code })
    throw error
  }
  return mode as ThemeMode
}

/**
 * Validate theme override colors per UX-THEME-001 contract.
 * Invalid colors are logged as warning and skipped (fall back to default).
 * @hlv THEME_OVERRIDE_INVALID_COLOR
 */
export function validateOverrides(overrides: ThemeOverride | undefined): ThemeOverride {
  if (!overrides) return {}

  const validOverrides: Record<string, string> = {}
  const colorKeys = ['primary', 'background', 'foreground']

  for (const [key, value] of Object.entries(overrides)) {
    if (colorKeys.includes(key) && !isValidHexColor(value)) {
      log.warn('theme.override.invalid_color', { key, value })
      continue // fall back to default, skip invalid
    }
    validOverrides[key] = value
  }

  return validOverrides as ThemeOverride
}

/**
 * Apply theme to document root.
 * Dark is default (:root), light is [data-theme="light"].
 */
export function applyThemeMode(mode: ThemeMode): void {
  if (mode === 'light') {
    document.documentElement.setAttribute('data-theme', 'light')
  } else {
    document.documentElement.removeAttribute('data-theme')
  }
  log.info('theme.mode.applied', { mode })
}

/**
 * Apply CSS variable overrides to document root.
 * Only overrides existing variables — never introduces new ones.
 * @hlv:sec [INPUT_VALIDATION] — overrides only applied to known CSS variable names
 */
export function applyOverrides(overrides: Record<string, string>): void {
  const root = document.documentElement
  const knownVariableMap: Record<string, string> = {
    primary: '--primary',
    background: '--background',
    foreground: '--foreground',
    font_family_sans: '--font-family-sans'
  }

  for (const [key, value] of Object.entries(overrides)) {
    const cssVar = knownVariableMap[key]
    if (cssVar) {
      root.style.setProperty(cssVar, value)
      log.info('theme.override.applied', { key, cssVar, value })
    }
  }
}

/**
 * Detect system theme preference.
 * Returns 'dark' if prefers-color-scheme: dark, otherwise 'light'.
 */
export function detectSystemPreference(): ThemeMode {
  if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
    return 'dark'
  }
  return 'light'
}

/**
 * Resolve theme configuration to response per UX-THEME-001 contract.
 * @hlv same_set_of_css_variables
 */
export function resolveTheme(config: ThemeConfig): ThemeResponse {
  const mode = validateMode(config.mode)
  void validateOverrides(config.overrides)

  // Collect all CSS variables from computed styles
  const cssVariables: Record<string, string> = {}
  const computed = getComputedStyle(document.documentElement)

  // Known CSS variable names — same set for both light and dark themes (UX-THEME-001 v2)
  const knownVars = [
    '--background',
    '--foreground',
    '--primary',
    '--primary-foreground',
    '--primary-hover',
    '--primary-muted',
    '--secondary',
    '--secondary-hover',
    '--accent',
    '--accent-hover',
    '--surface',
    '--surface-variant',
    '--error',
    '--error-bg',
    '--success',
    '--success-bg',
    '--warning',
    '--warning-bg',
    '--info',
    '--info-bg',
    '--border-default',
    '--border-focus',
    '--border-error',
    '--text-primary',
    '--text-secondary',
    '--text-muted',
    '--text-inverse',
    '--text-link',
    '--font-family-sans',
    '--font-family-mono',
    '--font-size-xs',
    '--font-size-sm',
    '--font-size-base',
    '--font-size-lg',
    '--font-size-xl',
    '--font-size-2xl',
    '--font-weight-normal',
    '--font-weight-medium',
    '--font-weight-semibold',
    '--font-weight-bold',
    '--line-height-tight',
    '--line-height-normal',
    '--line-height-relaxed',
    '--space-1',
    '--space-2',
    '--space-3',
    '--space-4',
    '--space-6',
    '--space-8',
    '--space-12',
    '--space-16',
    '--radius-sm',
    '--radius-md',
    '--radius-lg',
    '--radius-xl',
    '--radius-full',
    '--shadow-sm',
    '--shadow-md',
    '--shadow-lg',
    '--shadow-xl',
    '--transition-fast',
    '--transition-normal',
    '--transition-slow',
    '--z-dropdown',
    '--z-sticky',
    '--z-overlay',
    '--z-modal',
    '--z-toast',
    '--sidebar-width',
    '--sidebar-compact-width',
    '--header-height'
  ]

  for (const varName of knownVars) {
    const value = computed.getPropertyValue(varName).trim()
    if (value) {
      cssVariables[varName] = value
    }
  }

  return {
    mode,
    css_variables: cssVariables,
    theme_contract_version: THEME_CONTRACT_VERSION
  }
}
