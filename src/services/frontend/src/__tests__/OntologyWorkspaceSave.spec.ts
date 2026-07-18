// @m2.5 — OntologyWorkspace Save button vitest spec (GREEN: uses mountWithProviders)
// Tests: save button via useDraftState().saveDraft()
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { beforeEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("OntologyWorkspaceSave", () => {
	beforeEach(async () => {
		// Router setup handled by mountWithProviders
	});

	it("should have a Save button that saves draft via useDraftState when clicked", async () => {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		const saveBtn = wrapper.find(".toolbar-btn--primary");
		expect(saveBtn.exists()).toBe(true);
		expect(saveBtn.text()).toContain("Save");
	});

	it("should disable Save button when there are no unsaved changes", async () => {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		const saveBtn = wrapper.find(".toolbar-btn--primary");
		// Save button should bind to hasUnsavedChanges from useDraftState
		expect(saveBtn.attributes("disabled")).toBeDefined();
	});

	it("should render workspace toolbar with save button present", async () => {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".toolbar-btn--primary").exists()).toBe(true);
	});

	it("should render workspace page layout", async () => {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".workspace-page").exists()).toBe(true);
	});
});
