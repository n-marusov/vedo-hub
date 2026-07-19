const jwt = require('jsonwebtoken');
const fs = require('fs');

const privateKeyPem = fs.readFileSync('test-jwt-key.pem', 'utf8');

function signJWT(payload) {
  return jwt.sign(payload, privateKeyPem, {
    algorithm: 'RS256',
    expiresIn: '24h',
  });
}

// Owner token with admin roles — for CRUD operations
const ownerToken = signJWT({
  sub: 'user-123',
  user_id: 'user-123',
  tenant_id: '',
  roles: ['Owner'],
});

// Viewer token — for 403 tests (viewer trying to write)
const viewerToken = signJWT({
  sub: 'viewer-user',
  user_id: 'viewer-user',
  tenant_id: '',
  roles: ['Viewer'],
});

console.log(`OWNER_JWT=${ownerToken}`);
console.log(`VIEWER_JWT=${viewerToken}`);
