// @ctx: Global setup for Playwright API tests.
//
// 1. Auto-regenerates JWT tokens before test run when existing tokens are
//    expired or near expiry. This prevents the "expired token" regression
//    where hardcoded JWT timestamps fall behind the current date.
// 2. Cleans stale org test data (groups/projects from previous runs) so
//    consecutive runs start from a clean slate — without this, fixed-name
//    rows (TestGroup, TestProject, GUI Test Project, …) accumulate and break
//    strict-mode locators / duplicate-idempotency-key assertions.
//
// The tokens are signed with test-jwt-key.pem (RS256) and verified by the
// API gateway via JWT_DEV_PUBLIC_KEY_PEM in deploy/docker-compose.test.yml.
import { readFileSync, writeFileSync, existsSync } from 'fs';
import { resolve } from 'path';

// jsonwebtoken is an ESM-compatible CJS package; use dynamic import or require.
const jwt = require('jsonwebtoken');

const KEY_PATH = resolve(__dirname, 'test-jwt-key.pem');
const OUTPUT_PATH = resolve(__dirname, '..', 'specs', 'jwt-tokens.ts');
const API_BASE = 'http://localhost:3000/api/v1';

// Names of org rows created by E2E tests across runs. Anything matching these
// is deleted during global setup so the suite is repeatable.
const STALE_GROUP_NAMES = new Set([
  'TestGroup', 'UpdatedGroup', 'ParentGroup', 'ChildGroup', 'GUI-Test-Group',
  'US-CreateGroup', 'PairingGroup',
]);
const STALE_PROJECT_NAMES = new Set([
  'TestProject', 'UpdatedProject', 'GUI Test Project', 'US-CreateProject',
]);

/**
 * Decode the base64 payload of a JWT without verifying the signature.
 */
function decodePayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const raw = Buffer.from(parts[1], 'base64url').toString('utf8');
    return JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return null;
  }
}

/**
 * Check whether the existing jwt-tokens.ts contains tokens that are still
 * valid. Returns true when the OWNER_JWT token has an `exp` at least 1 hour
 * in the future, avoiding unnecessary file churn on every test run.
 */
function areTokensStillValid(): boolean {
  try {
    if (!existsSync(OUTPUT_PATH)) return false;
    const content = readFileSync(OUTPUT_PATH, 'utf8');
    const match = content.match(/OWNER_JWT\s*=\s*'([^']+)'/);
    if (!match) return false;
    const payload = decodePayload(match[1]);
    if (!payload || typeof payload.exp !== 'number') return false;
    const remainingMs = payload.exp * 1000 - Date.now();
    // Consider valid if more than 1 hour until expiry
    return remainingMs > 3600_000;
  } catch {
    return false;
  }
}

/**
 * Read the current OWNER_JWT from the generated tokens file (or sign a fresh
 * one if unavailable). Used to authenticate the cleanup requests.
 */
function getOwnerToken(privateKeyPem?: string): string {
  try {
    if (existsSync(OUTPUT_PATH)) {
      const content = readFileSync(OUTPUT_PATH, 'utf8');
      const match = content.match(/OWNER_JWT\s*=\s*'([^']+)'/);
      if (match) return match[1];
    }
  } catch {
    // fall through to signing a fresh token
  }
  if (privateKeyPem) {
    return jwt.sign(
      { sub: 'user-123', user_id: 'user-123', roles: ['Owner'] },
      privateKeyPem,
      { algorithm: 'RS256', expiresIn: '24h' },
    );
  }
  return '';
}

/**
 * List scopes (groups or projects) matching a name, then delete them.
 * Best-effort: failures are logged, not fatal — cleanup is a courtesy.
 */
async function cleanupScopes(resource: 'groups' | 'projects', names: Set<string>, token: string): Promise<void> {
  try {
    const listRes = await fetch(`${API_BASE}/${resource}?perPage=100`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!listRes.ok) {
      console.warn(`[setup] WARNING: could not list ${resource} for cleanup (HTTP ${listRes.status})`);
      return;
    }
    const body = await listRes.json() as { data?: Array<{ id?: string; name?: string; label?: string }> };
    const items = body.data ?? [];
    for (const item of items) {
      const name = item.name ?? item.label ?? '';
      if (!item.id || !names.has(name)) continue;
      const delRes = await fetch(`${API_BASE}/${resource}/${item.id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
          // Org write paths (incl. DELETE) require an Idempotency-Key.
          'Idempotency-Key': `cleanup-${item.id}-${Date.now()}`,
        },
      });
      if (delRes.ok || delRes.status === 404) {
        console.log(`[setup] cleanup: deleted stale ${resource.slice(0, -1)} "${name}" (${item.id})`);
      } else {
        console.warn(`[setup] WARNING: could not delete stale ${resource.slice(0, -1)} "${name}" (HTTP ${delRes.status})`);
      }
    }
  } catch (e) {
    console.warn('[setup] WARNING: cleanup failed', e);
  }
}

/**
 * Delete stale org test data created by previous runs.
 */
async function cleanupStaleData(privateKeyPem?: string): Promise<void> {
  const token = getOwnerToken(privateKeyPem);
  if (!token) {
    console.warn('[setup] WARNING: no owner token available — skipping org data cleanup');
    return;
  }
  // Projects first (they may reference groups as parents), then groups.
  await cleanupScopes('projects', STALE_PROJECT_NAMES, token);
  await cleanupScopes('groups', STALE_GROUP_NAMES, token);
}

/**
 * Global setup function — called once by Playwright before all tests.
 *
 * 1. Regenerates JWT tokens when expired/near-expiry.
 * 2. Deletes stale org test data so runs are repeatable.
 */
async function setup(): Promise<void> {
  let privateKeyPem: string | undefined;

  if (!existsSync(KEY_PATH)) {
    console.warn(
      '[setup] WARNING: test-jwt-key.pem not found at',
      KEY_PATH,
      '— JWT tokens will NOT be regenerated. Tests may fail with 401 if tokens expire.',
    );
  } else {
    privateKeyPem = readFileSync(KEY_PATH, 'utf8');
  }

  if (!privateKeyPem) {
    return;
  }

  if (areTokensStillValid()) {
    console.log('[setup] Existing JWT tokens are still valid — skipping regeneration.');
  } else {
    console.log('[setup] Existing JWT tokens expired or near expiry — regenerating...');

    function signToken(payload: Record<string, unknown>): string {
      return jwt.sign(payload, privateKeyPem as string, {
        algorithm: 'RS256',
        expiresIn: '24h',
      });
    }

    const ownerToken = signToken({
      sub: 'user-123',
      user_id: 'user-123',
      tenant_id: '',
      organization_id: 'org-001',
      roles: ['Owner'],
    });

    const viewerToken = signToken({
      sub: 'viewer-user',
      user_id: 'viewer-user',
      tenant_id: '',
      organization_id: 'org-001',
      roles: ['Viewer'],
    });

    const editorToken = signToken({
      sub: 'editor-user',
      user_id: 'editor-user',
      tenant_id: '',
      organization_id: 'org-001',
      roles: ['Editor'],
    });

    const generatedAt = new Date().toISOString();

    const content = [
      '// @ctx: M2.1/M2.5 E2E test JWT tokens — AUTO-GENERATED by global-setup.ts.',
      '// DO NOT EDIT manually — tokens are regenerated before test runs when needed.',
      '// Signed with test-jwt-key.pem (RS256). Verified by JWT_DEV_PUBLIC_KEY_PEM in',
      '// deploy/docker-compose.test.yml.',
      `// Generated: ${generatedAt} (expires in 24h)`,
      '',
      `export const OWNER_JWT = '${ownerToken}';`,
      '',
      `export const VIEWER_JWT = '${viewerToken}';`,
      '',
      `export const EDITOR_JWT = '${editorToken}';`,
      '',
    ].join('\n');

    writeFileSync(OUTPUT_PATH, content, 'utf8');
    console.log(`[setup] Fresh JWT tokens written → ${OUTPUT_PATH}`);
  }

  // Clean stale org data regardless of whether tokens were regenerated.
  await cleanupStaleData(privateKeyPem);
}

export default setup;
