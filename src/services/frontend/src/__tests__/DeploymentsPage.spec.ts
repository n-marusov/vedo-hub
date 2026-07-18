// @m2.5 — DeploymentsPage vitest spec (RED phase for Block В)
// After GREEN (Task 4.5): page should render deployment cards from API with show stopped toggle

import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { beforeEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("DeploymentsPage", () => {
	beforeEach(async () => {
		// Router setup handled by mountWithProviders
	});

	it("should render deployments page layout", async () => {
		const DeploymentsPage = (await import("@/pages/DeploymentsPage.vue"))
			.default;
		const wrapper = mountWithProviders(DeploymentsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dp-page").exists()).toBe(true);
	});

	it("should display page title Deployments", async () => {
		const DeploymentsPage = (await import("@/pages/DeploymentsPage.vue"))
			.default;
		const wrapper = mountWithProviders(DeploymentsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dp-page-title").exists()).toBe(true);
	});

	it("should render deployment cards from API data", async () => {
		const DeploymentsPage = (await import("@/pages/DeploymentsPage.vue"))
			.default;
		const wrapper = mountWithProviders(DeploymentsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dp-section").exists()).toBe(true);
	});

	it("should have show stopped deployments checkbox", async () => {
		const DeploymentsPage = (await import("@/pages/DeploymentsPage.vue"))
			.default;
		const wrapper = mountWithProviders(DeploymentsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dp-show-stopped").exists()).toBe(true);
	});

	it("should show breadcrumbs navigation", async () => {
		const DeploymentsPage = (await import("@/pages/DeploymentsPage.vue"))
			.default;
		const wrapper = mountWithProviders(DeploymentsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dp-breadcrumbs").exists()).toBe(true);
	});
});
