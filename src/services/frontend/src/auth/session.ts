// @ctx: auth session management — JWT storage, role extraction, token validation per GUI-LOGIN-001
// @hlv:sec [AUTH_BOUNDARY] — session token validated before granting route access
// @hlv:sec [SECRET_HANDLING] — JWT token stored in sessionStorage, not localStorage
export interface UserSession {
  accessToken: string
  refreshToken: string
  userId: string
  tenantId: string
  roles: string[]
  expiresAt: number
}

const SESSION_KEY = 'vedo_session'

// @hlv:sec [SECRET_HANDLING] — no credentials stored in localStorage after logout
export function saveSession(session: UserSession): void {
  sessionStorage.setItem(SESSION_KEY, JSON.stringify(session))
}

export function getSession(): UserSession | null {
  const raw = sessionStorage.getItem(SESSION_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as UserSession
  } catch {
    return null
  }
}

// @hlv GUI_AUTH_REQUIRED
export function isAuthenticated(): boolean {
  const session = getSession()
  if (!session) return false
  return isTokenValid()
}

// @hlv GUI_AUTH_REQUIRED
export function isTokenValid(): boolean {
  const session = getSession()
  if (!session) return false
  return Date.now() < session.expiresAt
}

// @ctx: role extraction for nav item visibility per GUI-OW-001 Security
export function getUserRole(): string | null {
  const session = getSession()
  if (!session || session.roles.length === 0) return null
  return session.roles[0]
}

export function hasRole(role: string): boolean {
  const session = getSession()
  if (!session) return false
  return session.roles.includes(role)
}

// @hlv:sec [AUTH_BOUNDARY] — logout clears all session data
export function logout(): void {
  sessionStorage.removeItem(SESSION_KEY)
}
