/**
 * Authenticated Playwright test fixture for GUI tests.
 *
 * Sets localStorage.vedo-jwt-token AND sessionStorage.vedo_session so the app's
 * initSession() (SKIP_AUTH mode) honors the injected session instead of seeding
 * the mock "skip-auth-token" (which the gateway rejects with 401).
 * Router auth is handled by VITE_SKIP_AUTH=true (build-time flag).
 */
import { test as base, expect } from '@playwright/test';
import { OWNER_JWT } from '../specs/jwt-tokens';

const INIT_SCRIPT = `(function() {
  const session = {
    accessToken: '${OWNER_JWT}',
    refreshToken: '${OWNER_JWT}',
    userId: 'user-123',
    tenantId: 'org-001',
    roles: ['Owner'],
    expiresAt: ${Date.now() + 86_400_000},
  };
  localStorage.setItem('vedo-jwt-token', '${OWNER_JWT}');
  sessionStorage.setItem('vedo_session', JSON.stringify(session));
})();`;

export const test = base.extend({
  context: async ({ context }, use) => {
    await context.addInitScript(INIT_SCRIPT);
    await use(context);
  },
});

export { expect };
