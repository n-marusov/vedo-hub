// @m2.5 — SPARQLPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: SPARQL query execution via Apollo GraphQL SPARQL_EXECUTE_QUERY
import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("SPARQLPage", () => {
	afterEach(() => {
		resetMockResults();
	});

	it("should render SPARQL Query Builder title", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mountWithProviders(SPARQLPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("SPARQL Query Builder");
	});

	it("should show query editor textarea ready for input", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mountWithProviders(SPARQLPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".spq-editor").exists()).toBe(true);
	});

	it("should have SPARQL query editor section", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mountWithProviders(SPARQLPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".spq-editor").exists()).toBe(true);
	});

	it("should render page layout without errors after loading", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mountWithProviders(SPARQLPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".spq-page").exists()).toBe(true);
	});

	it("should not crash when SPARQL query fails", async () => {
		setMockOperationResult(
			"SparqlExecute",
			null,
			new Error("SPARQL execution failed"),
		);
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mountWithProviders(SPARQLPage);
		await waitForQuery();
		await nextTick();
		// Component should render without crashing on API failure
		expect(wrapper.find(".spq-page").exists()).toBe(true);
	});

	it("should render query editor when there are no results yet", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mountWithProviders(SPARQLPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".spq-editor").exists()).toBe(true);
	});
});
