// Runtime configuration injected by docker-entrypoint.sh via /config.js.
// Values are non-sensitive (public URLs, feature flags, version metadata).
interface Window {
	__VEDO_CONFIG__?: {
		KEYCLOAK_URL?: string;
		KEYCLOAK_REALM?: string;
		KEYCLOAK_CLIENT_ID?: string;
		SKIP_AUTH?: string;
		GRAPHQL_ENDPOINT?: string;
		USE_MOCK_API?: string;
		APP_VERSION?: string;
		APP_ENV?: string;
		// Dev-only self-signed JWT minted on the host by `make docker-up`
		// (deploy/dev-jwt/gen-dev-jwt.js) and passed via VEDO_DEV_JWT_TOKEN.
		// Used by session.ts in SKIP_AUTH mode so API calls carry a
		// Gateway-accepted Bearer token.
		DEV_JWT_TOKEN?: string;
	};
}
