// @m4 — SPARQLPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: SPARQL query execution via Apollo GraphQL SPARQL_EXECUTE_QUERY
import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// Mock the direct apolloClient import used by SPARQLPage (it avoids the injection pattern)
vi.mock("@/apollo/client", () => {
	const mockQuery = vi.fn();
	return {
		apolloClient: { query: mockQuery },
	};
});

describePage("SPARQLPage", () => {
	afterEach(() => {
		resetMockResults();
		vi.clearAllMocks();
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

	it("should display GraphQL error message when server returns errors array", async () => {
		// Simulate GraphQL errors (not thrown) — the new code path added for
		// errorPolicy: 'all' which does not reject on GraphQL errors
		const { apolloClient } = await import("@/apollo/client");
		(apolloClient.query as ReturnType<typeof vi.fn>).mockResolvedValue({
			data: {},
			errors: [{ message: "Query syntax error at line 1" }],
		});

		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mountWithProviders(SPARQLPage);
		await waitForQuery();
		await nextTick();

		// Enter a query into the editor textarea
		const textarea = wrapper.find(".sparql-editor textarea");
		await textarea.setValue("SELECT INVALID");
		await nextTick();

		// Click the Run button — triggers onRunQuery which calls apolloClient.query()
		const buttons = wrapper.findAll("button");
		const runBtn = buttons.find((b) => /run/i.test(b.text()));
		if (!runBtn) throw new Error("Run button not found");
		await runBtn.trigger("click");
		await nextTick();

		// Wait for the async query to resolve
		await waitForQuery();
		await nextTick();

		// The component should display the error message from response.errors
		expect(wrapper.text()).toContain("Query syntax error at line 1");
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
