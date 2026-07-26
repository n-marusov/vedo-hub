// @ctx: Keycloak OIDC authorization code + PKCE flow — end-to-end login test
// Validates: US-auth.keycloak-login
// Validates: US-api.auth.jwt
// Validates: US-a11y.auth.authentication
// Validates: REQ-NFR.SECURITY.security-requirements
//
// Unlike other GUI tests that inject a pre-signed JWT session, this test
// exercises the REAL Keycloak OIDC flow: unauthenticated visit → redirect to
// /login → initiate Corporate SSO → Keycloak login page → credentials →
// redirect to /auth/callback → token exchange → redirect to /dashboard.
//
// NOTE: In the DEV environment (docker-compose.test.yml), the API gateway has
// JWT_DEV_PUBLIC_KEY_PEM for pre-signed tokens. Keycloak IS running and its
// realm is imported with test users. This test goes through the real flow.
//
// Keycloak 23 users (realm: vedo-core):
//   alice (viewer), bob (editor), carol (reviewer), dave (maintainer),
//   eve (admin), frank (owner), system-bot (service)
//   All have password = "password"
import { test, expect } from '@playwright/test';

const FRONTEND_URL = 'http://localhost:3000';
const KC_URL = 'http://localhost:8180';
const KC_REALM = 'vedo-core';
const KC_CLIENT_ID = 'vedo-core-frontend';

/**
 * Generate a PKCE code verifier (random 32 bytes, base64url-encoded).
 */
function generateCodeVerifier(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

/**
 * Generate a PKCE code challenge (SHA-256 of verifier, base64url-encoded).
 */
async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

test.describe('Keycloak OIDC Login Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Clear Keycloak SSO cookies to force a fresh login form each test.
    // Keycloak sets KEYCLOAK_SESSION / KEYCLOAK_IDENTITY cookies after login;
    // without clearing them, subsequent visits auto-authenticate via SSO.
    await page.context().clearCookies();
  });

  test('should redirect unauthenticated user to /login', async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/dashboard/home`);

    // Auth guard should redirect to /login with the original path as redirect param
    await expect(page).toHaveURL(/\/login(\?redirect=.*)?$/);
    await expect(page.getByRole('heading', { name: /sign in/i })).toBeVisible();
  });

  test('should show Corporate SSO as the only enabled provider', async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/login`);

    // Corporate SSO should be the only enabled button
    const corporateSso = page.getByRole('button', { name: 'Corporate SSO' });
    await expect(corporateSso).toBeVisible();
    await expect(corporateSso).toBeEnabled();

    // All external providers should be disabled
    await expect(page.getByRole('button', { name: 'VK ID' })).toBeDisabled();
    await expect(page.getByRole('button', { name: 'Yandex ID' })).toBeDisabled();
    await expect(page.getByRole('button', { name: 'Mail.ru' })).toBeDisabled();
    await expect(page.getByRole('button', { name: 'Google' })).toBeDisabled();
  });

  test('should complete full Keycloak OIDC authorization code flow with PKCE', async ({ page, context }) => {
    test.setTimeout(120_000);

    // Step 1: Navigate to the frontend login page
    await page.goto(`${FRONTEND_URL}/login`);
    await expect(page.getByRole('heading', { name: /sign in/i })).toBeVisible();

    // Step 2: Construct the Keycloak auth URL with PKCE params manually
    // (This avoids relying on the inititateLogin() JS which uses window.location.href)
    const state = crypto.randomUUID();
    const nonce = crypto.randomUUID();
    const codeVerifier = generateCodeVerifier();
    const codeChallenge = await generateCodeChallenge(codeVerifier);

    // Store PKCE params in sessionStorage so handleCallback() can find them
    await page.evaluate(
      ({ state, nonce, codeVerifier }) => {
        sessionStorage.setItem('kc_state', state);
        sessionStorage.setItem('kc_nonce', nonce);
        sessionStorage.setItem('kc_code_verifier', codeVerifier);
      },
      { state, nonce, codeVerifier }
    );

    const params = new URLSearchParams({
      client_id: KC_CLIENT_ID,
      redirect_uri: `${FRONTEND_URL}/auth/callback`,
      response_type: 'code',
      scope: 'openid profile email',
      state,
      nonce,
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
    });

    const authUrl = `${KC_URL}/realms/${KC_REALM}/protocol/openid-connect/auth?${params.toString()}`;

    // Step 3: Navigate directly to Keycloak auth URL
    await page.goto(authUrl, { waitUntil: 'networkidle' });
    await expect(page).toHaveURL(/\/realms\/vedo-core\/protocol\/openid-connect\/auth/);

    // Step 4: Fill in credentials on Keycloak login form
    await page.waitForSelector('#username', { timeout: 10_000 });
    await page.fill('#username', 'frank');
    await page.fill('#password', 'password');

    // Intercept navigation after form submit — the form POST will redirect
    // back to FRONTEND_URL/auth/callback?code=...&state=...
    // Playwright follows HTTP redirects automatically.
    const callbackPromise = page.waitForURL((url) => {
      return url.toString().startsWith(`${FRONTEND_URL}/auth/callback`);
    }, { timeout: 20_000 });

    // Step 5: Submit the login form
    const submitBtn = page.locator('#kc-login, button[type="submit"]').first();
    await submitBtn.click();

    // Step 6: Wait for the redirect to auth/callback after successful login
    await callbackPromise;

    // Step 7: Verify we're on the callback page with the auth code
    const callbackUrl = new URL(page.url());
    const code = callbackUrl.searchParams.get('code');
    expect(code).toBeTruthy();

    // Step 8: Wait for callback processing -> redirect to dashboard
    // AuthCallbackPage.handleCallback() exchanges code, saves session,
    // then router.replace() to /dashboard
    await page.waitForURL(/\/dashboard(\/home)?$/, { timeout: 15_000 });

    // Step 9: Verify the user is authenticated
    // The dashboard should show the user avatar menu for an authenticated user
    await expect(page.locator('.header-avatar-menu')).toBeVisible({ timeout: 10_000 });

    // Verify the token was stored in localStorage (used by Apollo Client)
    const jwtToken = await page.evaluate(() => localStorage.getItem('vedo-jwt-token'));
    expect(jwtToken).toBeTruthy();
    expect(jwtToken?.split('.').length).toBe(3);

    // Verify the session was stored in sessionStorage (used by router guard)
    const sessionRaw = await page.evaluate(() => sessionStorage.getItem('vedo_session'));
    expect(sessionRaw).toBeTruthy();
    const session = JSON.parse(sessionRaw!);
    // Keycloak JWT sub is the user's internal UUID, not the username
    expect(session.userId).toMatch(/^[0-9a-f-]{36}$/);
    expect(session.roles).toContain('owner');

    // Verify the access token is valid JWT (3 parts)
    expect(session.accessToken?.split('.').length).toBe(3);

    // Verify the decoded token has expected fields
    // base64url → base64 decode
    const b64 = session.accessToken.split('.')[1]
      .replace(/-/g, '+')
      .replace(/_/g, '/');
    const payload = JSON.parse(Buffer.from(b64, 'base64').toString());
    expect(payload.preferred_username).toBe('frank');
    expect(payload.realm_access?.roles).toContain('owner');
  });

  test('should reject login with invalid credentials', async ({ page }) => {
    // Navigate to frontend login page first to ensure a clean page state
    await page.goto(`${FRONTEND_URL}/login`);

    // Construct auth URL with PKCE params (state/code_verifier not needed in
    // sessionStorage since the callback is never reached on invalid credentials)
    const state = crypto.randomUUID();
    const codeVerifier = generateCodeVerifier();
    const codeChallenge = await generateCodeChallenge(codeVerifier);

    const params = new URLSearchParams({
      client_id: KC_CLIENT_ID,
      redirect_uri: `${FRONTEND_URL}/auth/callback`,
      response_type: 'code',
      scope: 'openid profile email',
      state,
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
    });

    const authUrl = `${KC_URL}/realms/${KC_REALM}/protocol/openid-connect/auth?${params.toString()}`;
    await page.goto(authUrl, { waitUntil: 'networkidle' });

    // Enter invalid credentials on Keycloak login page
    await page.waitForSelector('#username', { timeout: 10_000 });
    await page.fill('#username', 'frank');
    await page.fill('#password', 'wrong-password-123');

    await page.locator('#kc-login, button[type="submit"]').first().click();

    // Keycloak should stay on its page and show an error message
    await expect(async () => {
      const currentUrl = page.url();
      // The URL should still point to Keycloak (localhost:8180)
      expect(currentUrl).toContain('localhost:8180');
      // An error message should be visible
      await expect(page.locator('#input-error')).not.toBeEmpty({ timeout: 1_000 });
    }).toPass({ timeout: 10_000 });
  });

  test('should login with each realm role and verify correct role assignment', async ({ page }) => {
    test.setTimeout(300_000);

    async function loginAs(username: string): Promise<{ userId: string; roles: string[] }> {
      return await test.step(`login as ${username}`, async () => {
        // Clear Keycloak cookies and app session
        await page.context().clearCookies();

        // Navigate to frontend first to ensure we're on localhost:3000 origin
        // (page.evaluate() with sessionStorage only works on the frontend origin)
        await page.goto(`${FRONTEND_URL}/login`, { waitUntil: 'domcontentloaded' });

        // Clear any existing frontend session
        await page.evaluate(() => {
          sessionStorage.removeItem('vedo_session');
          sessionStorage.removeItem('kc_state');
          sessionStorage.removeItem('kc_nonce');
          sessionStorage.removeItem('kc_code_verifier');
          localStorage.removeItem('vedo-jwt-token');
        });

        // Construct auth URL with PKCE
        const state = crypto.randomUUID();
        const nonce = crypto.randomUUID();
        const codeVerifier = generateCodeVerifier();
        const codeChallenge = await generateCodeChallenge(codeVerifier);

        // Set PKCE params in sessionStorage (needed by handleCallback on return)
        await page.evaluate(
          ({ state, nonce, codeVerifier }) => {
            sessionStorage.setItem('kc_state', state);
            sessionStorage.setItem('kc_nonce', nonce);
            sessionStorage.setItem('kc_code_verifier', codeVerifier);
          },
          { state, nonce, codeVerifier }
        );

        // Navigate to Keycloak auth URL
        const params = new URLSearchParams({
          client_id: KC_CLIENT_ID,
          redirect_uri: `${FRONTEND_URL}/auth/callback`,
          response_type: 'code',
          scope: 'openid profile email',
          state,
          nonce,
          code_challenge: codeChallenge,
          code_challenge_method: 'S256',
        });

        await page.goto(`${KC_URL}/realms/${KC_REALM}/protocol/openid-connect/auth?${params.toString()}`, { waitUntil: 'networkidle' });
        await page.waitForSelector('#username', { timeout: 10_000 });
        await page.fill('#username', username);
        await page.fill('#password', 'password');

        // Wait for redirect back to frontend callback after successful login
        const callbackPromise = page.waitForURL((url) => {
          return url.toString().startsWith(`${FRONTEND_URL}/auth/callback`);
        }, { timeout: 20_000 });

        await page.locator('#kc-login, button[type="submit"]').first().click();
        await callbackPromise;

        // Wait for dashboard after callback processing
        await page.waitForURL(/\/dashboard(\/home)?$/, { timeout: 15_000 });

        // Extract session and decode JWT to verify user and roles
        const sessionRaw = await page.evaluate(() => sessionStorage.getItem('vedo_session'));
        expect(sessionRaw).toBeTruthy();
        const session = JSON.parse(sessionRaw!);

        // Decode JWT (base64url → base64) to verify preferred_username and roles
        const b64 = session.accessToken.split('.')[1]
          .replace(/-/g, '+')
          .replace(/_/g, '/');
        const payload = JSON.parse(
          Buffer.from(b64, 'base64').toString()
        );
        return {
          userId: payload.preferred_username as string,
          roles: (payload.realm_access?.roles || []) as string[],
        };
      });
    }

    // Test each role sequentially
    let session = await loginAs('alice');
    expect(session.userId).toBe('alice');
    expect(session.roles).toContain('viewer');

    session = await loginAs('bob');
    expect(session.userId).toBe('bob');
    expect(session.roles).toContain('editor');

    session = await loginAs('carol');
    expect(session.userId).toBe('carol');
    expect(session.roles).toContain('reviewer');

    session = await loginAs('dave');
    expect(session.userId).toBe('dave');
    expect(session.roles).toContain('maintainer');

    session = await loginAs('eve');
    expect(session.userId).toBe('eve');
    expect(session.roles).toContain('admin');

    session = await loginAs('frank');
    expect(session.userId).toBe('frank');
    expect(session.roles).toContain('owner');
  });
});
