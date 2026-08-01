#!/bin/sh
# =============================================================================
# docker-entrypoint.sh — Minimal entrypoint for publish-browse-ui.
# =============================================================================
# SECURITY: config.js is public. APP_VERSION is the only exposed value.
# NEVER add secrets here.
# =============================================================================

cat > /usr/share/nginx/html/config.js <<SAFE_EOF
window.__VEDO_CONFIG__ = { APP_VERSION: "${VEDO_APP_VERSION:-dev}" };
SAFE_EOF

exec nginx -g "daemon off;"
