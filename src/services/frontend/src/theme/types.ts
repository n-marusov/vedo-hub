// @hlv:artifact code-frontend implements spec-ux-theme-001
// @ctx: theme types — ThemeMode, ThemeConfig, ThemeOverride per UX-THEME-001 contract

export type ThemeMode = 'light' | 'dark'

export interface ThemeOverride {
  primary?: string
  background?: string
  foreground?: string
  font_family_sans?: string
}

export interface ThemeConfig {
  mode: ThemeMode
  overrides?: ThemeOverride
}

export interface ThemeResponse {
  mode: ThemeMode
  css_variables: Record<string, string>
  theme_contract_version: string
}

export const VALID_THEME_MODES: ThemeMode[] = ['light', 'dark']
export const THEME_CONTRACT_VERSION = '2.0.0'

// @hlv:sec [INPUT_VALIDATION] — hex color validation for theme overrides
const HEX_COLOR_REGEX = /^#[0-9a-fA-F]{6}$/

export function isValidHexColor(value: string): boolean {
  return HEX_COLOR_REGEX.test(value)
}

export const ERROR_CODES = {
  THEME_INVALID_MODE: 'THEME_INVALID_MODE',
  THEME_OVERRIDE_INVALID_COLOR: 'THEME_OVERRIDE_INVALID_COLOR'
} as const

export class ThemeError extends Error {
  constructor(
    public code: keyof typeof ERROR_CODES,
    message: string
  ) {
    super(message)
    this.name = 'ThemeError'
  }
}
