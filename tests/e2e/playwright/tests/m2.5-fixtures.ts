// @ctx: M2.5 E2E test fixtures — mock auth, API route proxying
import { test as base, expect } from '@playwright/test';

// Create a mock JWT and session for E2E testing without Keycloak
function createMockSession() {
  const payload = {
    sub: 'user-123',
    name: 'Test User',
    preferred_username: 'owner_seed',
    email: 'owner@vedo.dev',
    tenant_id: 'default',
    realm_access: { roles: ['owner', 'editor'] },
    exp: Math.floor(Date.now() / 1000) + 86400,
  };
  const encoded = btoa(JSON.stringify(payload));
  const mockToken = `header.${encoded}.signature`;

  return {
    accessToken: mockToken,
    refreshToken: mockToken,
    userId: 'user-123',
    tenantId: 'default',
    roles: ['owner'],
    expiresAt: Date.now() + 86400000,
  };
}

export const test = base.extend({
  page: async ({ page }, use) => {
    const session = createMockSession();
    await page.addInitScript((s) => {
      // Set auth session for router guard (sessionStorage key: vedo_session)
      sessionStorage.setItem('vedo_session', JSON.stringify({
        accessToken: s.accessToken,
        refreshToken: s.refreshToken,
        userId: s.userId,
        tenantId: s.tenantId,
        roles: s.roles,
        expiresAt: s.expiresAt,
      }));

      // Set JWT for useCurrentUser composable (localStorage key: vedo-jwt-token)
      const userPayload = {
        sub: s.userId,
        name: 'Test User',
        preferred_username: 'owner_seed',
        email: 'owner@vedo.dev',
        picture: '',
        exp: Math.floor(Date.now() / 1000) + 86400,
      };
      const encoded = btoa(JSON.stringify(userPayload));
      localStorage.setItem('vedo-jwt-token', `header.${encoded}.signature`);
    }, session);

    await use(page);
  },
});

export { expect } from '@playwright/test';
