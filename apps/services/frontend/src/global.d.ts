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
  };
}
