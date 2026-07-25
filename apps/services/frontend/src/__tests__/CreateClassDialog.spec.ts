// @m4 — CreateClassDialog vitest spec
// Validates: REQ-USR.UI.gui-implementation
//
// @aif — Migrated from Apollo `createClass` mutation to REST `@/api/ontology`
// per ADR-DES.API.rest-graphql-mutation-boundary.md.
import {
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import CreateClassDialog from "@/components/ontology/CreateClassDialog.vue";
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

// Teleport stub: renders slot inline instead of moving to document.body
const TeleportStub = { template: "<div><slot /></div>" };

describe("CreateClassDialog", () => {
	it("should render when open is true", async () => {
		const wrapper = mountWithProviders(CreateClassDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dialog-overlay").exists()).toBe(true);
		expect(wrapper.text()).toContain("Create Class");
	});

	it("should not render when open is false", async () => {
		const wrapper = mountWithProviders(CreateClassDialog, {
			props: { open: false, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		expect(wrapper.find(".dialog-overlay").exists()).toBe(false);
	});

	it("should show validation error when name is empty", async () => {
		const wrapper = mountWithProviders(CreateClassDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const createBtn = wrapper.find(".btn--primary");
		expect(createBtn.exists()).toBe(true);
		await createBtn.trigger("click");
		await nextTick();
		expect(wrapper.text()).toContain("Class name is required");
	});

	it("should emit created event on successful submit", async () => {
		const wrapper = mountWithProviders(CreateClassDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const input = wrapper.find(".form-input");
		await input.setValue("Person");
		const createBtn = wrapper.find(".btn--primary");
		await createBtn.trigger("click");
		await new Promise((resolve) => setTimeout(resolve, 600));
		expect(wrapper.emitted("created")).toBeTruthy();
		expect(wrapper.emitted("created")?.[0]).toEqual(["Person"]);
	});

	it("should emit close on cancel", async () => {
		const wrapper = mountWithProviders(CreateClassDialog, {
			props: { open: true, ontologyId: "test-onto-id" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const cancelBtn = wrapper
			.findAll("button")
			.filter((b) => b.text().includes("Cancel"));
		expect(cancelBtn.length).toBeGreaterThanOrEqual(1);
		await cancelBtn[0]?.trigger("click");
		expect(wrapper.emitted("close")).toBeTruthy();
	});
});
