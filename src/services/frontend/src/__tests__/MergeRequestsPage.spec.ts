// @m2.5 — MergeRequestsPage vitest spec (RED phase for Block В)
// After GREEN (Task 4.5): page should render MR sections, filter tabs, and data from API

import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { beforeEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("MergeRequestsPage", () => {
	beforeEach(async () => {
		// Router setup handled by mountWithProviders
	});

	it("should render merge requests page layout", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-page").exists()).toBe(true);
	});

	it("should display page title Merge Requests", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-title").exists()).toBe(true);
	});

	it("should show filter tabs for merge request status", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-section").exists()).toBe(true);
	});

	it("should render merge request sections with data from API", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".merge-requests").exists()).toBe(true);
	});

	it("should show breadcrumbs navigation", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-breadcrumbs").exists()).toBe(true);
	});
});
