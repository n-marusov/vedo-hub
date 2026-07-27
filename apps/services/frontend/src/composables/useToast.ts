// @hlv:artifact code-frontend toast-notification
// @ctx: Global Toast notification composable — singleton state for show/dismiss

import { ref } from "vue";

export interface ToastOptions {
	message: string;
	type?: "success" | "error";
	duration?: number;
}

const toastMessage = ref("");
const toastType = ref<"success" | "error">("success");
const toastDuration = ref(5000);
const toastVisible = ref(false);

let toastTimer: ReturnType<typeof setTimeout> | null = null;

function showToast(message: string, type: "success" | "error" = "success") {
	if (toastTimer !== null) {
		clearTimeout(toastTimer);
		toastTimer = null;
	}

	toastMessage.value = message;
	toastType.value = type;
	toastVisible.value = true;

	console.info(
		JSON.stringify({
			event: "useToast.show",
			type,
			message,
		}),
	);

	toastTimer = setTimeout(() => {
		dismissToast();
	}, toastDuration.value);
}

function dismissToast() {
	if (toastTimer !== null) {
		clearTimeout(toastTimer);
		toastTimer = null;
	}

	toastVisible.value = false;

	console.info(
		JSON.stringify({
			event: "useToast.dismiss",
		}),
	);
}

export function useToast() {
	return {
		toastMessage,
		toastType,
		toastVisible,
		showToast,
		dismissToast,
	};
}
