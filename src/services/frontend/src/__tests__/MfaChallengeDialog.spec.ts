// @m2.5 — MfaChallengeDialog vitest spec
import {
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import MfaChallengeDialog from "@/components/auth/MfaChallengeDialog.vue";
import { describe, expect, it } from "vitest";
import { nextTick } from "vue";

const TeleportStub = { template: "<div><slot /></div>" };

describe("MfaChallengeDialog", () => {
	it("should render when open is true", async () => {
		const wrapper = mountWithProviders(MfaChallengeDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dialog-overlay").exists()).toBe(true);
		expect(wrapper.text()).toContain("Two-Factor Authentication");
	});

	it("should have disabled Verify button when code is empty", async () => {
		const wrapper = mountWithProviders(MfaChallengeDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const verifyBtn = wrapper.find(".btn--primary");
		expect(verifyBtn.attributes("disabled")).toBeDefined();
	});

	it("should enable Verify button with 6-digit code", async () => {
		const wrapper = mountWithProviders(MfaChallengeDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const input = wrapper.find(".form-input--code");
		await input.setValue("123456");
		await nextTick();
		const verifyBtn = wrapper.find(".btn--primary");
		expect(verifyBtn.attributes("disabled")).toBeUndefined();
	});

	it("should emit verified on successful verification", async () => {
		const wrapper = mountWithProviders(MfaChallengeDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const input = wrapper.find(".form-input--code");
		await input.setValue("123456");
		await nextTick();
		const verifyBtn = wrapper.find(".btn--primary");
		await verifyBtn.trigger("click");
		await new Promise((resolve) => setTimeout(resolve, 1000));
		expect(wrapper.emitted("verified")).toBeTruthy();
	});

	it("should show resend link with cooldown", async () => {
		const wrapper = mountWithProviders(MfaChallengeDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const resendBtn = wrapper.find(".resend-link");
		expect(resendBtn.exists()).toBe(true);
		expect(resendBtn.text()).toContain("Resend");
	});
});
