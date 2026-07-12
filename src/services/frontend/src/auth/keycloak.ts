// Keycloak OIDC client — authorization code flow with PKCE per GUI-LOGIN-001

import { type UserSession, saveSession } from './session'

export interface KeycloakConfig {
  realm: string
  clientId: string
  url: string
  redirectUri: string
}

interface JwtPayload {
  sub?: string
  tenant_id?: string
  realm_access?: { roles?: string[] }
  [key: string]: unknown
}

const DEFAULT_CONFIG: KeycloakConfig = {
  realm: 'vedo-core',
  clientId: 'vedo-spa',
  url: 'http://127.0.0.1:8443',
  redirectUri: `${window.location.origin}/auth/callback`
}

export function getConfig(): KeycloakConfig {
  return {
    realm: import.meta.env.VITE_KEYCLOAK_REALM || DEFAULT_CONFIG.realm,
    clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || DEFAULT_CONFIG.clientId,
    url: import.meta.env.VITE_KEYCLOAK_URL || DEFAULT_CONFIG.url,
    redirectUri: DEFAULT_CONFIG.redirectUri
  }
}

function generateCodeVerifier(): string {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder()
  const data = encoder.encode(verifier)
  const digest = await crypto.subtle.digest('SHA-256', data)
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
}

export async function initiateLogin(): Promise<void> {
  const config = getConfig()
  const state = crypto.randomUUID()
  const nonce = crypto.randomUUID()
  const codeVerifier = generateCodeVerifier()

  sessionStorage.setItem('kc_state', state)
  sessionStorage.setItem('kc_nonce', nonce)
  sessionStorage.setItem('kc_code_verifier', codeVerifier)

  const codeChallenge = await generateCodeChallenge(codeVerifier)

  const params = new URLSearchParams({
    client_id: config.clientId,
    redirect_uri: config.redirectUri,
    response_type: 'code',
    scope: 'openid profile email',
    state,
    nonce,
    code_challenge: codeChallenge,
    code_challenge_method: 'S256'
  })

  const authUrl = `${config.url}/realms/${config.realm}/protocol/openid-connect/auth?${params.toString()}`
  window.location.href = authUrl
}

export async function handleCallback(): Promise<UserSession> {
  const params = new URLSearchParams(window.location.search)
  const code = params.get('code')
  const state = params.get('state')
  const error = params.get('error')

  if (error) {
    const errorDesc = params.get('error_description') || error
    throw new Error(`OIDC error: ${errorDesc}`)
  }

  if (!code || !state) {
    throw new Error('Missing authorization code or state')
  }

  const savedState = sessionStorage.getItem('kc_state')
  if (state !== savedState) {
    throw new Error('State mismatch — possible CSRF attack')
  }

  const config = getConfig()
  const codeVerifier = sessionStorage.getItem('kc_code_verifier')
  if (!codeVerifier) {
    throw new Error('Missing code verifier')
  }

  const tokenResponse = await fetch(
    `${config.url}/realms/${config.realm}/protocol/openid-connect/token`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        client_id: config.clientId,
        code,
        redirect_uri: config.redirectUri,
        code_verifier: codeVerifier
      })
    }
  )

  if (!tokenResponse.ok) {
    const body = await tokenResponse.text()
    throw new Error(`Token exchange failed: ${body}`)
  }

  const tokens = await tokenResponse.json()
  const session = parseToken(tokens.access_token, tokens.refresh_token)

  sessionStorage.removeItem('kc_state')
  sessionStorage.removeItem('kc_nonce')
  sessionStorage.removeItem('kc_code_verifier')

  saveSession(session)
  return session
}

function parseToken(accessToken: string, refreshToken: string): UserSession {
  const payload = decodeJwtPayload(accessToken) as JwtPayload

  const roles: string[] = (payload.realm_access?.roles || []).filter(
    (r: string) => !r.startsWith('default-') && r !== 'offline_access' && r !== 'uma_authorization'
  )

  return {
    accessToken,
    refreshToken,
    userId: payload.sub as string,
    tenantId: (payload.tenant_id as string) || 'default',
    roles,
    expiresAt: (payload.exp as number) * 1000
  }
}

function decodeJwtPayload(token: string): Record<string, unknown> {
  const parts = token.split('.')
  if (parts.length !== 3) {
    throw new Error('Invalid JWT token')
  }
  const padded = parts[1]
    .replace(/-/g, '+')
    .replace(/_/g, '/')
    .padEnd(parts[1].length + ((4 - (parts[1].length % 4)) % 4), '=')
  const json = decodeURIComponent(atob(padded))
  return JSON.parse(json)
}

export function logout(): void {
  const config = getConfig()
  const endSessionUrl = `${config.url}/realms/${config.realm}/protocol/openid-connect/logout`
  window.location.href = endSessionUrl
}
