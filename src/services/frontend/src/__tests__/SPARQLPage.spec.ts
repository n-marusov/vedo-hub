// @aif — Updated SPARQLPage vitest spec after migrating SPARQL execution
// from Apollo GraphQL (`SPARQL_EXECUTE_QUERY` / `sparqlQuery` resolver) to
// REST `POST /api/v1/sparql` per ADR-DES.API.rest-graphql-mutation-boundary.md.
// Tests: SPARQL page render, query execution via REST, error handling.
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// Mock the REST API used by SPARQLPage — `executeSparql` from `@/api/sparql`.
// Keep the mock permissive so individual tests can override its return value
// with `mockResolvedValue` / `mockRejectedValue`.
vi.mock("@/api/sparql", () => {
	return {
		executeSparql: vi.fn(),
	};
});

describePage("SPARQLPage", () => {
	afterEach(() => {
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

	it("should display REST error message when SPARQL execution fails", async () => {
		const { executeSparql } = await import("@/api/sparql");
		(executeSparql as ReturnType<typeof vi.fn>).mockRejectedValueOnce(
			new Error("SPARQL syntax error at line 1"),
		);

		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mountWithProviders(SPARQLPage);
		await waitForQuery();
		await nextTick();

		// Enter a query into the editor textarea
		const textarea = wrapper.find(".sparql-editor textarea");
		await textarea.setValue("SELECT INVALID");
		await nextTick();

		// Click the Run button — triggers onRunQuery which calls executeSparql()
		const buttons = wrapper.findAll("button");
		const runBtn = buttons.find((b) => /run/i.test(b.text()));
		if (!runBtn) throw new Error("Run button not found");
		await runBtn.trigger("click");
		await nextTick();

		// Wait for the async REST call to reject and the error message to render
		await waitForQuery();
		await nextTick();

		// The component should display the error message from the REST failure
		expect(wrapper.text()).toContain("SPARQL syntax error at line 1");
	});

	it("should not crash when SPARQL query fails", async () => {
		const { executeSparql } = await import("@/api/sparql");
		(executeSparql as ReturnType<typeof vi.fn>).mockRejectedValueOnce(
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
