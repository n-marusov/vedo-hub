import { mount } from "@vue/test-utils";
// @m2.5 — OntologyWorkspace Save button vitest spec (RED phase: will fail on hardcoded save)
// After GREEN (Task 2.3): save button should use useDraftState().saveDraft()
import { beforeEach, describe, expect, it } from "vitest";
import { nextTick } from "vue";
import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{
			path: "/ontology/:id/workspace",
			name: "ontology-workspace",
			component: { template: "<div />" },
		},
	],
});

describe("OntologyWorkspace Save", () => {
	beforeEach(async () => {
		await router.push("/ontology/test/workspace");
	});

	it("should have a Save button that saves draft via useDraftState when clicked", async () => {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mount(OntologyWorkspace, {
			global: { plugins: [router] },
		});
		await nextTick();
		const saveBtn = wrapper.find(".toolbar-btn--primary");
		expect(saveBtn.exists()).toBe(true);
		expect(saveBtn.text()).toContain("Save");
	});

	it("should disable Save button when there are no unsaved changes", async () => {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mount(OntologyWorkspace, {
			global: { plugins: [router] },
		});
		await nextTick();
		// RED: Save button should bind to hasUnsavedChanges
		const saveBtn = wrapper.find(".toolbar-btn--primary");
		expect(saveBtn.attributes("disabled")).toBeDefined();
	});

	it("should show loading state while save is in progress", async () => {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mount(OntologyWorkspace, {
			global: { plugins: [router] },
		});
		await nextTick();
		// RED: Should show loading state during save
		const saveBtn = wrapper.find(".toolbar-btn--primary");
		expect(saveBtn.exists()).toBe(true);
	});

	it("should show error state when save fails", async () => {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mount(OntologyWorkspace, {
			global: { plugins: [router] },
		});
		await nextTick();
		// RED: Should show error state when useDraftState().saveDraft() returns false
		expect(wrapper.find(".workspace-error").exists()).toBe(false);
	});
});
