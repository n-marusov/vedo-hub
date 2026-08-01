// Dev-mode runtime config — used when running `vite` directly (not Docker).
// In Docker, this file is overwritten by docker-entrypoint.sh from VEDO_* env vars.
// SECURITY: this file is only active in local development, never in production.
window.__VEDO_CONFIG__ = {
	KEYCLOAK_URL: "http://localhost:8180",
	KEYCLOAK_REALM: "vedo-core",
	KEYCLOAK_CLIENT_ID: "vedo-core-frontend",
	SKIP_AUTH: "true",
	GRAPHQL_ENDPOINT: "/api/v1/graphql",
	USE_MOCK_API: "false",
	APP_VERSION: "dev",
	APP_ENV: "dev",
};
