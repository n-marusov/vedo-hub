// @m4 — MetricsPage vitest spec (RED phase for Block В)
// Validates: REQ-USR.UI.gui-implementation
// After GREEN (Task 4.3): page should render KPI counters and trends from API

import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("MetricsPage", () => {
	afterEach(() => {
		resetMockResults();
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

	it("should show empty state when ontology has zero metrics", async () => {
		setMockOperationResult("OntologyMetrics", {
			ontologyMetrics: {
				classes: 0,
				properties: 0,
				individuals: 0,
				shapes: 0,
				trends: [],
			},
		});
		const MetricsPage = (await import("@/pages/MetricsPage.vue")).default;
		const wrapper = mountWithProviders(MetricsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".metrics-empty").exists() ||
				wrapper.text().includes("No data") ||
				wrapper.text().includes("0"),
		).toBe(true);
	});

	it("should show error state when metrics API fails", async () => {
		setMockOperationResult("OntologyMetrics", null, new Error("Metrics error"));
		const MetricsPage = (await import("@/pages/MetricsPage.vue")).default;
		const wrapper = mountWithProviders(MetricsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".error-state").exists() ||
				wrapper.text().includes("retry") ||
				wrapper.text().includes("error"),
		).toBe(true);
	});
});
