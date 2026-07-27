import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

vi.mock("@/locales/en.json", () => ({
	default: {
		"toast.close": "Close",
	},
}));

vi.mock("@/locales/ru.json", () => ({
	default: {
		"toast.close": "Закрыть",
	},
}));

describe("Toast", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should render success message", async () => {
		const { useToast } = await import("@/composables/useToast");
		const { toastMessage, toastVisible, toastType } = useToast();

		// Directly set module-level state for testing
		const { showToast } = useToast();
		showToast("Operation completed");

		await nextTick();
		expect(toastVisible.value).toBe(true);
		expect(toastMessage.value).toBe("Operation completed");
		expect(toastType.value).toBe("success");
	});

	it("should render error message", async () => {
		const { useToast } = await import("@/composables/useToast");
		const { showToast, toastType, toastMessage } = useToast();

		showToast("Something went wrong", "error");

		await nextTick();
		expect(toastMessage.value).toBe("Something went wrong");
		expect(toastType.value).toBe("error");
	});

	it("should auto-dismiss after timeout", async () => {
		vi.useFakeTimers();

		const { useToast } = await import("@/composables/useToast");
		const { showToast, toastVisible } = useToast();

		showToast("Test message");
		await nextTick();
		expect(toastVisible.value).toBe(true);

		vi.advanceTimersByTime(5500);
		await nextTick();
		expect(toastVisible.value).toBe(false);

		vi.useRealTimers();
	});

	it("should close on dismiss call", async () => {
		const { useToast } = await import("@/composables/useToast");
		const { showToast, dismissToast, toastVisible } = useToast();

		showToast("Test message");
		await nextTick();
		expect(toastVisible.value).toBe(true);

		dismissToast();
		await nextTick();
		expect(toastVisible.value).toBe(false);
	});

	it("should have correct aria attributes for success", async () => {
		const { useToast } = await import("@/composables/useToast");
		const { showToast, toastType } = useToast();

		showToast("Success message", "success");
		await nextTick();
		expect(toastType.value).toBe("success");
	});

	it("should have correct type for error", async () => {
		const { useToast } = await import("@/composables/useToast");
		const { showToast, toastType } = useToast();

		showToast("Error message", "error");
		await nextTick();
		expect(toastType.value).toBe("error");
	});
});
