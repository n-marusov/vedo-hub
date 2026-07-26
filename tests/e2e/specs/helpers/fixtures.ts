import { test as base } from '@playwright/test';

// Pre-signed RS256 JWT with Owner role — matches api-gateway-test's public key
import { OWNER_JWT } from './jwt-tokens';

type SeededUser = {
  username: string;
  role: 'owner' | 'editor' | 'viewer';
};

function createMockSession() {
  return {
    accessToken: OWNER_JWT,
    refreshToken: OWNER_JWT,
    userId: 'user-123',
    tenantId: 'default',
    roles: ['owner'],
    expiresAt: Date.now() + 86400000,
  };
}

export const test = base.extend<{ seededUsers: SeededUser[] }>({
  seededUsers: async ({}, use) => {
    await use([
      { username: 'owner_seed', role: 'owner' },
      { username: 'editor_seed', role: 'editor' },
      { username: 'viewer_seed', role: 'viewer' }
    ]);
  },

  page: async ({ page }, use) => {
    const session = createMockSession();
    const ownerJwt = OWNER_JWT;

    await page.addInitScript(({ session, ownerJwt }) => {
      // Set auth session for router guard
      sessionStorage.setItem('vedo_session', JSON.stringify(session));
      // Set JWT for Apollo Client and REST API
      localStorage.setItem('vedo-jwt-token', ownerJwt);
    }, { session, ownerJwt });

    await use(page);
  },
});

export { expect } from '@playwright/test';
