import { mount } from "@vue/test-utils";
// @m2.5 — SPARQLPage vitest spec (RED phase: will fail on hardcoded query execution)
// After GREEN (Task 2.4): query execution should use Apollo GraphQL SPARQL_EXECUTE_QUERY
import { beforeEach, describe, expect, it } from "vitest";
import { nextTick } from "vue";
import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{
			path: "/ontology/:id/query",
			name: "ontology-query",
			component: { template: "<div />" },
		},
	],
});

describe("SPARQLPage", () => {
	beforeEach(async () => {
		await router.push("/ontology/test/query");
	});

	it("should execute SPARQL query via Apollo when Run is clicked", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mount(SPARQLPage, {
			global: { plugins: [router] },
		});
		await nextTick();
		// RED: Should execute SPARQL query via GraphQL
		expect(wrapper.text()).toContain("SPARQL Query Builder");
	});

	it("should display results table when query execution succeeds", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mount(SPARQLPage, {
			global: { plugins: [router] },
		});
		await nextTick();
		// RED: Should show results table after successful query execution
		expect(wrapper.find(".query-results-table").exists()).toBe(false);
	});

	it("should show error message when query execution fails", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mount(SPARQLPage, {
			global: { plugins: [router] },
		});
		await nextTick();
		// RED: Should display error from API
		expect(wrapper.find(".spq-error").exists()).toBe(false);
	});

	it("should format query when format button is clicked", async () => {
		const SPARQLPage = (await import("@/pages/SPARQLPage.vue")).default;
		const wrapper = mount(SPARQLPage, {
			global: { plugins: [router] },
		});
		await nextTick();
		// RED: Format should trigger formatting endpoint
		expect(wrapper.text()).toContain("SPARQL Query Builder");
	});
});
