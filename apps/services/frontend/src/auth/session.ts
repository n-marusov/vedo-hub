// auth session management — JWT storage, role extraction, token validation per GUI-LOGIN-001
export interface UserSession {
	accessToken: string;
	refreshToken: string;
	userId: string;
	tenantId: string;
	roles: string[];
	expiresAt: number;
}

// SKIP_AUTH is disabled in production builds — Vite's dead-code elimination
// removes all SKIP_AUTH branches at build time when MODE=production.
// In non-production builds (dev server, dev/e2e images), the runtime
// window.__VEDO_CONFIG__.SKIP_AUTH value is honored.
// NOTE: must use MODE (not import.meta.env.PROD) — vite build sets
// NODE_ENV=production (→ PROD=true) for ANY --mode, so PROD would
// tree-shake SKIP_AUTH even in dev/e2e builds.
const SKIP_AUTH =
	import.meta.env.MODE === "production"
		? false
		: (window.__VEDO_CONFIG__?.SKIP_AUTH || "false") === "true";

const SESSION_KEY = "vedo_session";

// Mock session for test mode — provides a valid session so Apollo Client
// and Axios interceptors can inject Authorization header without Keycloak.
// When the dev overlay mints a real self-signed JWT (DEV_JWT_TOKEN in
// window.__VEDO_CONFIG__), that token is used instead so the API Gateway
// accepts the request (skip-auth-token alone is rejected with 401).
const MOCK_SESSION: UserSession = {
	accessToken: "skip-auth-token",
	refreshToken: "skip-auth-token",
	userId: "test-user",
	tenantId: "org-001",
	roles: ["Owner"],
	expiresAt: 9999999999999,
};

// Decode the exp claim (seconds) from a JWT payload without verifying the
// signature — used only to seed expiresAt for the dev-minted token.
function decodeJwtExpiry(token: string): number {
	try {
		const payload = token.split(".")[1];
		if (!payload) return Number.MAX_SAFE_INTEGER;
		const padded = payload
			.replace(/-/g, "+")
			.replace(/_/g, "/")
			.padEnd(payload.length + ((4 - (payload.length % 4)) % 4), "=");
		const json = JSON.parse(decodeURIComponent(atob(padded)));
		return typeof json.exp === "number"
			? json.exp * 1000
			: Number.MAX_SAFE_INTEGER;
	} catch {
		return Number.MAX_SAFE_INTEGER;
	}
}

// @ctx: store access token also in localStorage.vedo-jwt-token so Apollo Client
// and Axios interceptors can read it. The auth flow stores the full session in
// sessionStorage (vedo_session) but GraphQL/REST clients read from localStorage.
export function saveSession(session: UserSession): void {
	sessionStorage.setItem(SESSION_KEY, JSON.stringify(session));
	localStorage.setItem("vedo-jwt-token", session.accessToken);
}

// Eagerly seed the session at app startup (called once from main.ts).
// In SKIP_AUTH mode the router guard short-circuits without calling
// getSession(), so without this the dev-minted JWT would never reach
// localStorage.vedo-jwt-token — API clients (Apollo/Axios) read that key
// directly and would send no (or a stale) Authorization header → 401.
// Always calls getSession() (even when a session already exists) so the
// token is synced to localStorage in every SKIP_AUTH branch.
export function initSession(): void {
	if (SKIP_AUTH) {
		getSession();
	}
}

export function getSession(): UserSession | null {
	// In SKIP_AUTH (test) mode, honor an explicitly injected session first
	// (e.g. tests that set sessionStorage.vedo_session to control roles),
	// then the dev-minted JWT (DEV_JWT_TOKEN), then the mock session.
	if (SKIP_AUTH) {
		const raw = sessionStorage.getItem(SESSION_KEY);
		if (raw) {
			try {
				const parsed = JSON.parse(raw) as UserSession;
				// Keep localStorage in sync so API clients carry the token.
				if (parsed.accessToken) {
					localStorage.setItem("vedo-jwt-token", parsed.accessToken);
				}
				return parsed;
			} catch {
				// fall through to dev token / mock
			}
		}
		const devToken = window.__VEDO_CONFIG__?.DEV_JWT_TOKEN;
		if (devToken) {
			// Save via saveSession to sync both storage backends — API clients
			// (Apollo, Axios) read localStorage.vedo-jwt-token for the header.
			const devSession: UserSession = {
				accessToken: devToken,
				refreshToken: devToken,
				userId: "dev-user",
				tenantId: "org-001",
				roles: ["Owner"],
				expiresAt: decodeJwtExpiry(devToken),
			};
			saveSession(devSession);
			return devSession;
		}
		// No dev token (e.g. e2e without the dev overlay): seed the mock
		// session token too so interceptors still attach an Authorization
		// header (the gateway is configured to accept only the dev key in
		// dev/test, but the browser must at least send SOMETHING consistent).
		if (!sessionStorage.getItem(SESSION_KEY)) {
			saveSession(MOCK_SESSION);
		}
		return MOCK_SESSION;
	}
	const raw = sessionStorage.getItem(SESSION_KEY);
	if (!raw) return null;
	try {
		return JSON.parse(raw) as UserSession;
	} catch {
		return null;
	}
}

export function isAuthenticated(): boolean {
	if (SKIP_AUTH) return true;
	const session = getSession();
	if (!session) return false;
	return isTokenValid();
}

export function isTokenValid(): boolean {
	if (SKIP_AUTH) return true;
	const session = getSession();
	if (!session) return false;
	return Date.now() < session.expiresAt;
}

// role extraction for nav item visibility per GUI-OW-001 Security
export function getUserRole(): string | null {
	const session = getSession();
	if (!session || session.roles.length === 0) return null;
	return session.roles[0];
}

export function hasRole(role: string): boolean {
	const session = getSession();
	if (!session) return false;
	return session.roles.includes(role);
}

export function logout(): void {
	sessionStorage.removeItem(SESSION_KEY);
	localStorage.removeItem("vedo-jwt-token");
}
