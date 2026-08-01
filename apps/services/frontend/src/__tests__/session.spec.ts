// @m5 — Regression tests for dev-auth session seeding (initSession).
//
// Background: in SKIP_AUTH (dev/test) mode the router guard short-circuits
// without calling getSession(), so the dev-minted JWT (DEV_JWT_TOKEN from
// window.__VEDO_CONFIG__) only reached localStorage.vedo-jwt-token lazily —
// when some component happened to call getSession(). API clients (Apollo,
// Axios) read localStorage directly, so org write endpoints returned 401.
// initSession() seeds the token eagerly at app startup (main.ts).
//
// Regression: if initSession() is removed or the SKIP_AUTH mode stops
// seeding localStorage, dev API calls break with 401.
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const DEV_TOKEN = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.dev.payload'
const SESSION_KEY = 'vedo_session'
const TOKEN_KEY = 'vedo-jwt-token'

function clearStores(): void {
  sessionStorage.clear()
  localStorage.clear()
}

function setConfig(cfg: Record<string, string> | undefined): void {
  Object.defineProperty(window, '__VEDO_CONFIG__', {
    value: cfg,
    writable: true,
    configurable: true
  })
}

describe('session.ts — initSession (SKIP_AUTH dev overlay)', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('import.meta', { env: { MODE: 'dev' } })
    clearStores()
    setConfig({ SKIP_AUTH: 'true', DEV_JWT_TOKEN: DEV_TOKEN })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    clearStores()
  })

  it('seeds localStorage.vedo-jwt-token with the dev JWT in SKIP_AUTH mode', async () => {
    const { initSession } = await import('@/auth/session')
    initSession()

    expect(localStorage.getItem(TOKEN_KEY)).toBe(DEV_TOKEN)
    // The full session is stored so role extraction works too.
    const raw = sessionStorage.getItem(SESSION_KEY)
    expect(raw).not.toBeNull()
    const session = JSON.parse(raw as string)
    expect(session.accessToken).toBe(DEV_TOKEN)
    expect(session.roles).toContain('Owner')
  })

  it('does not overwrite an explicitly injected session (tests control roles)', async () => {
    sessionStorage.setItem(
      SESSION_KEY,
      JSON.stringify({
        accessToken: 'injected-token',
        refreshToken: 'injected-token',
        userId: 'u1',
        tenantId: 'org-001',
        roles: ['Viewer'],
        expiresAt: 9999999999999
      })
    )
    const { initSession } = await import('@/auth/session')
    initSession()

    expect(localStorage.getItem(TOKEN_KEY)).toBe('injected-token')
    expect(JSON.parse(sessionStorage.getItem(SESSION_KEY) as string).roles).toEqual(['Viewer'])
  })

  it('is idempotent — repeated calls keep the first session', async () => {
    const { initSession } = await import('@/auth/session')
    initSession()
    const first = sessionStorage.getItem(SESSION_KEY)
    initSession()
    expect(sessionStorage.getItem(SESSION_KEY)).toBe(first)
  })

  it('does not seed when no dev token is configured (falls back to mock session)', async () => {
    setConfig({ SKIP_AUTH: 'true' })
    const { initSession } = await import('@/auth/session')
    initSession()
    // Mock session token is seeded so interceptors still attach a header.
    expect(localStorage.getItem(TOKEN_KEY)).toBe('skip-auth-token')
  })

  it('does nothing in non-SKIP_AUTH mode', async () => {
    setConfig({ SKIP_AUTH: 'false', DEV_JWT_TOKEN: DEV_TOKEN })
    const { initSession } = await import('@/auth/session')
    initSession()
    expect(localStorage.getItem(TOKEN_KEY)).toBeNull()
    expect(sessionStorage.getItem(SESSION_KEY)).toBeNull()
  })
})
