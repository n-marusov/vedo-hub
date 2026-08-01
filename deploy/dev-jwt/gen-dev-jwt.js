#!/usr/bin/env node
/**
 * gen-dev-jwt.js — mints a fresh self-signed dev JWT on the host.
 *
 * Dev analogue of tests/e2e/scripts/global-setup.ts: generates a JWT at
 * `make docker-up` (ENV=dev) time, signed with the shared dev/test private
 * key (tests/e2e/scripts/test-jwt-key.pem, RS256). The API Gateway accepts
 * it because the dev overlay sets JWT_DEV_PUBLIC_KEY_PEM (the matching
 * public key) on the api-gateway service.
 *
 * Output: prints a single JWT on stdout. The Makefile passes it to the
 * frontend container as VEDO_DEV_JWT_TOKEN, which the entrypoint copies
 * into config.js → window.__VEDO_CONFIG__.DEV_JWT_TOKEN → session.ts seeds
 * localStorage.vedo-jwt-token so API calls carry a valid Bearer token.
 *
 * Usage (from repo root):
 *   node deploy/dev-jwt/gen-dev-jwt.js
 */
const { createRequire } = require('module');
const fs = require('fs');
const path = require('path');

// jsonwebtoken is installed in tests/e2e (used by the E2E token tooling);
// resolve it by absolute path so this script runs from any cwd.
const requireFromE2E = createRequire(
  path.join(__dirname, '..', '..', 'tests', 'e2e', 'scripts', 'noop.js'),
);
const jwt = requireFromE2E('jsonwebtoken');

const PRIVATE_KEY_PATH = path.join(
  __dirname,
  '..',
  '..',
  'tests',
  'e2e',
  'scripts',
  'test-jwt-key.pem',
);

// Mirror the E2E Owner token claims (see tests/e2e/scripts/sign-jwt.js) so the
// gateway role gate (Owner weight 3 >= POST Editor level 1) passes for org
// write endpoints and auth-service sees a consistent identity.
const OWNER_CLAIMS = {
  sub: 'user-123',
  user_id: 'user-123',
  tenant_id: '',
  organization_id: 'org-001',
  roles: ['Owner'],
};

function main() {
  if (!fs.existsSync(PRIVATE_KEY_PATH)) {
    process.stderr.write(
      `[gen-dev-jwt] FATAL: dev signing key not found at ${PRIVATE_KEY_PATH}\n`,
    );
    process.exit(1);
  }

  const privateKeyPem = fs.readFileSync(PRIVATE_KEY_PATH, 'utf8');

  const token = jwt.sign(OWNER_CLAIMS, privateKeyPem, {
    algorithm: 'RS256',
    expiresIn: '24h',
  });

  process.stdout.write(token + '\n');
  process.stderr.write(
    '[gen-dev-jwt] dev JWT minted (subject=user-123, roles=Owner, exp=24h)\n',
  );
}

main();
