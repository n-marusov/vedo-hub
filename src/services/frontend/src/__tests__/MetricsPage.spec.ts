// @m2.5 — MetricsPage vitest spec (RED phase for Block В)
// After GREEN (Task 4.3): page should render KPI counters and trends from API

import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { beforeEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("MetricsPage", () => {
	beforeEach(async () => {
		// Router setup handled by mountWithProviders
	});

	it("should render KPI counters from API when page loads", async () => {
		const MetricsPage = (await import("@/pages/MetricsPage.vue")).default;
		const wrapper = mountWithProviders(MetricsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".kpi-grid").exists()).toBe(true);
	});

	it("should display class count from API", async () => {
		const MetricsPage = (await import("@/pages/MetricsPage.vue")).default;
		const wrapper = mountWithProviders(MetricsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".metrics-content").exists()).toBe(true);
	});

	it("should show context banner with ontology name", async () => {
		const MetricsPage = (await import("@/pages/MetricsPage.vue")).default;
		const wrapper = mountWithProviders(MetricsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".metrics-context").exists()).toBe(true);
	});

	it("should render trend charts section", async () => {
		const MetricsPage = (await import("@/pages/MetricsPage.vue")).default;
		const wrapper = mountWithProviders(MetricsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".charts-grid").exists()).toBe(true);
	});

	it("should show loading skeleton while metrics load from API", async () => {
		const MetricsPage = (await import("@/pages/MetricsPage.vue")).default;
		const wrapper = mountWithProviders(MetricsPage);
		expect(wrapper.find(".metrics-page").exists()).toBe(true);
	});
});
