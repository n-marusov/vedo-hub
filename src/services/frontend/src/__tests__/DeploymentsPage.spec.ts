// @m4 — DeploymentsPage vitest spec (RED phase for Block В)
// Validates: REQ-USR.UI.gui-implementation
// After GREEN (Task 4.5): page should render deployment cards from API with show stopped toggle

import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("DeploymentsPage", () => {
	afterEach(() => {
		resetMockResults();
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

	it("should show error state when deployments API fails", async () => {
		setMockOperationResult(
			"ListDeployments",
			null,
			new Error("Failed to load deployments"),
		);
		const DeploymentsPage = (await import("@/pages/DeploymentsPage.vue"))
			.default;
		const wrapper = mountWithProviders(DeploymentsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".error-state").exists() ||
				wrapper.text().includes("retry") ||
				wrapper.text().includes("error"),
		).toBe(true);
	});

	it("should show empty state when no deployments exist", async () => {
		setMockOperationResult("ListDeployments", {
			deployments: [],
		});
		const DeploymentsPage = (await import("@/pages/DeploymentsPage.vue"))
			.default;
		const wrapper = mountWithProviders(DeploymentsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".dp-empty").exists() ||
				wrapper.find(".empty-state").exists() ||
				wrapper.text().includes("No deployments"),
		).toBe(true);
	});
});
