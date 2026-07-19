// @m2.5 — useCurrentUser composable for App.vue and DashboardPage user display
// Decodes JWT from Keycloak session stored in localStorage, exposes name/email/initials

import { computed, onMounted, ref } from "vue";

interface CurrentUser {
	name: string;
	email: string;
	avatarUrl: string;
	initials: string;
}

const user = ref<CurrentUser>({
	name: "",
	email: "",
	avatarUrl: "",
	initials: "?",
});

function decodeJwt(token: string): Record<string, unknown> | null {
	try {
		const base64Url = token.split(".")[1];
		const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
		const jsonPayload = decodeURIComponent(
			atob(base64)
				.split("")
				.map((c) => `%${(`00${c.charCodeAt(0).toString(16)}`).slice(-2)}`)
				.join(""),
		);
		return JSON.parse(jsonPayload);
	} catch {
		return null;
	}
}

export function useCurrentUser() {
	onMounted(() => {
		const token = localStorage.getItem("vedo-jwt-token");
		if (!token) return;

		const payload = decodeJwt(token);
		if (!payload) return;

		const name =
			(payload.name as string) || (payload.preferred_username as string) || "";
		const email = (payload.email as string) || "";
		const avatarUrl = (payload.picture as string) || "";

		// Generate initials from name
		const initials = name
			.split(" ")
			.map((n) => n[0]?.toUpperCase() || "")
			.join("")
			.slice(0, 2);

		user.value = { name, email, avatarUrl, initials: initials || "?" };
	});

	const displayName = computed(() => user.value.name || "User");
	const displayInitials = computed(() => user.value.initials);

	return {
		user,
		displayName,
		displayInitials,
	};
}
