// @m5 — OntologyWorkspace three-view navigation spec (Q3).
// Validates: REQ-USR.UI.gui-implementation
// Verifies the Class Hierarchy / TBox Graph / ABox Graph navigation split:
// the switcher renders three tabs, the ClassTree organism renders in
// hierarchy view, and switching views shows the corresponding panel.
import {
	describePage,
	mountWithProviders,
	resetMockResults,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// OntologyWorkspace uses axios.get() directly for fetchOntologyMeta(), and
// src/api/ai.ts calls axios.create() + interceptors at module load — the mock
// must provide a full instance (same shape as OntologyWorkspaceSave.spec.ts).
vi.mock("axios", () => {
	const mockInstance = {
		get: vi.fn().mockResolvedValue({
			data: {
				id: "ont-1",
				name: "Test Ontology",
				branch: "main",
				dirty: false,
			},
		}),
		post: vi.fn().mockResolvedValue({ data: {} }),
		put: vi.fn().mockResolvedValue({ data: {} }),
		delete: vi.fn().mockResolvedValue({ data: {} }),
		create: vi.fn().mockReturnThis(),
		interceptors: {
			request: { use: vi.fn() },
			response: { use: vi.fn() },
		},
	};
	return {
		default: mockInstance,
		create: vi.fn().mockReturnValue(mockInstance),
		get: mockInstance.get,
		post: mockInstance.post,
	};
});

describePage("OntologyWorkspaceViews", () => {
	afterEach(() => {
		resetMockResults();
		vi.clearAllMocks();
	});

	async function mountWorkspace() {
		const OntologyWorkspace = (await import("@/pages/OntologyWorkspace.vue"))
			.default;
		const wrapper = mountWithProviders(OntologyWorkspace, {
			global: {
				stubs: {
					CreateClassDialog: true,
					CreatePropertyDialog: true,
					CreateIndividualDialog: true,
					GraphVisualization: true,
					AiSuggestionPanel: true,
				},
			},
		});
		await waitForQuery();
		await nextTick();
		return wrapper;
	}

	it("should render the three-view navigation switcher", async () => {
		const wrapper = await mountWorkspace();
		const buttons = wrapper.findAll(".nav-switcher__btn");
		expect(buttons.length).toBe(3);
		// Hierarchy is the default active view.
		expect(wrapper.find(".nav-switcher__btn--active").text()).toContain(
			"Class Hierarchy",
		);
	});

	it("should render the ClassTree organism in hierarchy view", async () => {
		const wrapper = await mountWorkspace();
		// Default view is hierarchy — ClassTree renders the class panel.
		expect(wrapper.find(".class-tree").exists()).toBe(true);
		// ClassTree renders the mocked class "owl:Thing".
		expect(wrapper.text()).toContain("owl:Thing");
	});

	it("should switch to TBox graph view when TBox tab is clicked", async () => {
		const wrapper = await mountWorkspace();
		const tboxBtn = wrapper
			.findAll(".nav-switcher__btn")
			.find((b) => b.text().includes("TBox Graph"));
		if (!tboxBtn) throw new Error("TBox Graph tab not found");
		await tboxBtn.trigger("click");
		await nextTick();
		// TBox view shows the GraphVisualization stub; ClassTree is hidden.
		expect(wrapper.find(".class-tree").exists()).toBe(false);
		expect(wrapper.findComponent({ name: "GraphVisualization" }).exists()).toBe(
			true,
		);
	});

	it("should switch to ABox graph view when ABox tab is clicked", async () => {
		const wrapper = await mountWorkspace();
		const aboxBtn = wrapper
			.findAll(".nav-switcher__btn")
			.find((b) => b.text().includes("ABox Graph"));
		if (!aboxBtn) throw new Error("ABox Graph tab not found");
		await aboxBtn.trigger("click");
		await nextTick();
		// ABox view shows the individuals table header.
		expect(wrapper.find(".graph-head").exists()).toBe(true);
		expect(wrapper.find(".class-tree").exists()).toBe(false);
	});
});
