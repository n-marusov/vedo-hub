// auth session management — JWT storage, role extraction, token validation per GUI-LOGIN-001
export interface UserSession {
	accessToken: string;
	refreshToken: string;
	userId: string;
	tenantId: string;
	roles: string[];
	expiresAt: number;
}

const SKIP_AUTH = import.meta.env.VITE_SKIP_AUTH === "true";

const SESSION_KEY = "vedo_session";

// Mock session for test mode — provides a valid session so Apollo Client
// and Axios interceptors can inject Authorization header without Keycloak.
const MOCK_SESSION: UserSession = {
	accessToken: "skip-auth-token",
	refreshToken: "skip-auth-token",
	userId: "test-user",
	tenantId: "org-001",
	roles: ["Owner"],
	expiresAt: 9999999999999,
};

// @ctx: store access token also in localStorage.vedo-jwt-token so Apollo Client
// and Axios interceptors can read it. The auth flow stores the full session in
// sessionStorage (vedo_session) but GraphQL/REST clients read from localStorage.
export function saveSession(session: UserSession): void {
	sessionStorage.setItem(SESSION_KEY, JSON.stringify(session));
	localStorage.setItem("vedo-jwt-token", session.accessToken);
}

export function getSession(): UserSession | null {
	if (SKIP_AUTH) return MOCK_SESSION;
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
