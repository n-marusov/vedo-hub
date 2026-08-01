/**
 * Authenticated Playwright test fixture for GUI tests.
 *
 * Sets localStorage.vedo-jwt-token so Apollo/Axios interceptors
 * can inject Authorization header for API calls.
 * Router auth is handled by VITE_SKIP_AUTH=true (build-time flag).
 */
import { test as base, expect } from '@playwright/test';
import { OWNER_JWT } from '../specs/jwt-tokens';

const INIT_SCRIPT = `(function() {
  localStorage.setItem('vedo-jwt-token', '${OWNER_JWT}');
})();`;

export const test = base.extend({
  context: async ({ context }, use) => {
    await context.addInitScript(INIT_SCRIPT);
    await use(context);
  },
});

export { expect };
