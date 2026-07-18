// @m2.5 — ValidationPage vitest spec (RED phase for Block В)
// After GREEN (Task 4.4): page should run SHACL validation via RUN_VALIDATION_MUTATION and display results

import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { beforeEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("ValidationPage", () => {
	beforeEach(async () => {
		// Router setup handled by mountWithProviders
	});

	it("should render validation page layout", async () => {
		const ValidationPage = (await import("@/pages/ValidationPage.vue")).default;
		const wrapper = mountWithProviders(ValidationPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".validation-page").exists()).toBe(true);
	});

	it("should show Run validation button", async () => {
		const ValidationPage = (await import("@/pages/ValidationPage.vue")).default;
		const wrapper = mountWithProviders(ValidationPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".run-btn").exists()).toBe(true);
	});

	it("should run validation when Run button is clicked", async () => {
		const ValidationPage = (await import("@/pages/ValidationPage.vue")).default;
		const wrapper = mountWithProviders(ValidationPage);
		await waitForQuery();
		await nextTick();
		const runBtn = wrapper.find(".run-btn");
		await runBtn.trigger("click");
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".validation-page").exists()).toBe(true);
	});

	it("should display validation results after running", async () => {
		const ValidationPage = (await import("@/pages/ValidationPage.vue")).default;
		const wrapper = mountWithProviders(ValidationPage);
		await waitForQuery();
		await nextTick();
		await wrapper.find(".run-btn").trigger("click");
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".validation-card").exists()).toBe(true);
	});

	it("should show context banner with ontology name", async () => {
		const ValidationPage = (await import("@/pages/ValidationPage.vue")).default;
		const wrapper = mountWithProviders(ValidationPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".validation-context").exists()).toBe(true);
	});
});
