// @ctx: structured logging for router navigation per observability constraints
export interface NavigationLog {
  route: string
  path: string
  timestamp: string
}

export interface AuthRedirectLog {
  reason: string
  target: string
  timestamp: string
}

export function logNavigation(data: { route: string; path: string }): void {
  console.info(
    JSON.stringify({
      level: 'info',
      event: 'router.navigation',
      route: data.route,
      path: data.path,
      timestamp: new Date().toISOString()
    })
  )
}

export function logAuthRedirect(data: { reason: string; target: string }): void {
  console.info(
    JSON.stringify({
      level: 'warn',
      event: 'router.auth_redirect',
      reason: data.reason,
      target: data.target,
      timestamp: new Date().toISOString()
    })
  )
}
