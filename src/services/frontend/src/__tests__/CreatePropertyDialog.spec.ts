// @m4 — CreatePropertyDialog vitest spec
// Validates: REQ-USR.UI.gui-implementation
//
// @aif — Migrated from Apollo `createProperty` mutation to REST
// `@/api/ontology` per ADR-DES.API.rest-graphql-mutation-boundary.md.
import {
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import CreatePropertyDialog from "@/components/ontology/CreatePropertyDialog.vue";
import { describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// Mock REST ontology API — no HTTP in jsdom.
vi.mock("@/api/ontology", () => ({
	createClass: vi.fn().mockResolvedValue({
		id: "new-class-1",
		label: "Person",
		comment: null,
		parents: [] as string[],
		children: [] as string[],
	}),
	createProperty: vi.fn().mockResolvedValue({
		id: "new-property-1",
		label: "hasName",
		propertyType: "object",
		domains: [] as string[],
		ranges: [] as string[],
	}),
	createIndividual: vi.fn().mockResolvedValue({
		id: "new-individual-1",
		label: "JohnDoe",
		classId: "owl:Thing",
		classLabel: "owl:Thing",
	}),
}));

const TeleportStub = { template: "<div><slot /></div>" };

describe("CreatePropertyDialog", () => {
	it("should render when open is true", async () => {
		const wrapper = mountWithProviders(CreatePropertyDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dialog-overlay").exists()).toBe(true);
		expect(wrapper.text()).toContain("Create Property");
	});

	it("should show config tab by default", async () => {
		const wrapper = mountWithProviders(CreatePropertyDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		expect(wrapper.text()).toContain("Property Name");
		expect(wrapper.text()).toContain("Property Type");
	});

	it("should switch to preview tab", async () => {
		const wrapper = mountWithProviders(CreatePropertyDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const tabs = wrapper.findAll(".tab");
		const previewTab = tabs[1];
		await previewTab?.trigger("click");
		await nextTick();
		expect(wrapper.find(".turtle-preview").exists()).toBe(true);
	});

	it("should validate empty name", async () => {
		const wrapper = mountWithProviders(CreatePropertyDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const createBtn = wrapper.find(".btn--primary");
		await createBtn.trigger("click");
		await nextTick();
		expect(wrapper.text()).toContain("Property name is required");
	});

	it("should emit created on successful submit", async () => {
		const wrapper = mountWithProviders(CreatePropertyDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const input = wrapper.find(".form-input");
		await input.setValue("hasName");
		const createBtn = wrapper.find(".btn--primary");
		await createBtn.trigger("click");
		await new Promise((resolve) => setTimeout(resolve, 600));
		expect(wrapper.emitted("created")).toBeTruthy();
		expect(wrapper.emitted("created")?.[0]).toEqual(["hasName"]);
	});
});
