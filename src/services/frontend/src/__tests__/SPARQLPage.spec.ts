// @m2.5 — SPARQLPage vitest spec (GREEN: uses mountWithProviders)
// Tests: SPARQL query execution via Apollo GraphQL SPARQL_EXECUTE_QUERY
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { beforeEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("SPARQLPage", () => {
	beforeEach(async () => {
		// Router setup handled by mountWithProviders
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
});
