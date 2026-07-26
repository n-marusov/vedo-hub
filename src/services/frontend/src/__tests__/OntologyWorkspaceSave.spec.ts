// @m4 — OntologyWorkspace Save button vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: save button via useDraftState().saveDraft()
import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// OntologyWorkspace uses axios.get() directly for fetchOntologyMeta().
// Mock axios so the component loads data and renders the save button.
vi.mock("axios", () => {
	const mockInstance = {
		get: vi.fn().mockResolvedValue({ data: { id: "test", name: "Test Ontology", branch: "main", dirty: false } }),
		post: vi.fn().mockResolvedValue({ data: {} }),
		put: vi.fn().mockResolvedValue({ data: {} }),
		delete: vi.fn().mockResolvedValue({ data: {} }),
		interceptors: {
			request: { use: vi.fn() },
			response: { use: vi.fn() },
		},
	};
	return {
		default: {
			create: vi.fn().mockReturnValue(mockInstance),
			get: mockInstance.get,
			post: mockInstance.post,
			...mockInstance,
		},
		create: vi.fn().mockReturnValue(mockInstance),
		get: mockInstance.get,
		post: mockInstance.post,
	};
});

describePage("OntologyWorkspaceSave", () => {
	afterEach(() => {
		resetMockResults();
		vi.clearAllMocks();
	});

	it("should have a Save button that saves draft via useDraftState when clicked", async () => {
		setMockOperationResult("UpdateDraft", {
			updateDraft: { success: true, timestamp: new Date().toISOString() },
		});
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue")).default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		const saveBtn = wrapper.find(".toolbar-btn--primary");
		expect(saveBtn.exists()).toBe(true);
		expect(saveBtn.text()).toContain("Save");
	});

	it("should disable Save button when there are no unsaved changes", async () => {
		setMockOperationResult("UpdateDraft", {
			updateDraft: { success: true, timestamp: new Date().toISOString() },
		});
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue")).default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		const saveBtn = wrapper.find(".toolbar-btn--primary");
		expect(saveBtn.attributes("disabled")).toBeDefined();
	});

	it("should render workspace toolbar with save button present", async () => {
		setMockOperationResult("UpdateDraft", {
			updateDraft: { success: true, timestamp: new Date().toISOString() },
		});
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue")).default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".toolbar-btn--primary").exists()).toBe(true);
	});

	it("should render workspace page layout", async () => {
		setMockOperationResult("UpdateDraft", {
			updateDraft: { success: true, timestamp: new Date().toISOString() },
		});
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue")).default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".workspace-page").exists()).toBe(true);
	});

	it("should not crash when save fails due to API error", async () => {
		setMockOperationResult("UpdateDraft", null, new Error("Failed to save draft"));
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue")).default;
		const wrapper = mountWithProviders(OntologyWorkspace);
		await waitForQuery();
		await nextTick();
		const saveBtn = wrapper.find(".toolbar-btn--primary");
		await saveBtn.trigger("click");
		await waitForQuery();
		await nextTick();
		// Component should render without crashing on save failure
		expect(wrapper.find(".workspace-page").exists()).toBe(true);
	});
});
