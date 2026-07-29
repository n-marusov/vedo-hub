// REST API client for user preferences.
//
// After GraphQL tightening: preferences migrated from GraphQL
// (NAVIGATION_STATE_QUERY) to localStorage-based client.

// ── Types ──────────────────────────────────────────────────────────────────────────

export interface UserPreferences {
	sidebarCollapsed: boolean;
	activeRoute: string;
	theme: "light" | "dark" | "system";
}

const STORAGE_KEY = "vedo-user-preferences";

const DEFAULTS: UserPreferences = {
	sidebarCollapsed: false,
	activeRoute: "/",
	theme: "system",
};

// ── Get Preferences ───────────────────────────────────────────────────────────────

export function getUserPreferences(): UserPreferences {
	console.info(
		JSON.stringify({
			level: "info",
			msg: "preferences.get.request",
			ts: new Date().toISOString(),
		}),
	);

	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (raw) {
			const parsed = JSON.parse(raw) as Partial<UserPreferences>;
			const prefs = { ...DEFAULTS, ...parsed };
			console.info(
				JSON.stringify({
					level: "info",
					msg: "preferences.get.success",
					theme: prefs.theme,
					ts: new Date().toISOString(),
				}),
			);
			return prefs;
		}
	} catch (err: unknown) {
		const e = err as { message?: string };
		console.warn(
			JSON.stringify({
				level: "warn",
				msg: "preferences.get.parse_error",
				error: e.message,
				ts: new Date().toISOString(),
			}),
		);
	}

	return { ...DEFAULTS };
}

// ── Update Preferences ────────────────────────────────────────────────────────────

export function updateUserPreferences(
	partial: Partial<UserPreferences>,
): UserPreferences {
	console.info(
		JSON.stringify({
			level: "info",
			msg: "preferences.update.request",
			...partial,
			ts: new Date().toISOString(),
		}),
	);

	const current = getUserPreferences();
	const updated: UserPreferences = { ...current, ...partial };

	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
		console.info(
			JSON.stringify({
				level: "info",
				msg: "preferences.update.success",
				ts: new Date().toISOString(),
			}),
		);
	} catch (err: unknown) {
		const e = err as { message?: string };
		console.error(
			JSON.stringify({
				level: "error",
				msg: "preferences.update.failed",
				error: e.message,
				ts: new Date().toISOString(),
			}),
		);
	}

	return updated;
}
