// @m4 — CreateIndividualDialog vitest spec
// Validates: REQ-USR.UI.gui-implementation
//
// @aif — Migrated from Apollo `createIndividual` mutation to REST
// `@/api/ontology` per ADR-DES.API.rest-graphql-mutation-boundary.md.
import {
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import CreateIndividualDialog from "@/components/ontology/CreateIndividualDialog.vue";
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

describe("CreateIndividualDialog", () => {
	it("should render when open is true", async () => {
		const wrapper = mountWithProviders(CreateIndividualDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dialog-overlay").exists()).toBe(true);
		expect(wrapper.text()).toContain("Create Individual");
	});

	it("should validate empty name", async () => {
		const wrapper = mountWithProviders(CreateIndividualDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const createBtn = wrapper.find(".btn--primary");
		await createBtn.trigger("click");
		await nextTick();
		expect(wrapper.text()).toContain("Individual name is required");
	});

	it("should add property rows dynamically", async () => {
		const wrapper = mountWithProviders(CreateIndividualDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		// Find add button by text
		const addBtn = wrapper
			.findAll("button")
			.filter((b) => b.text().includes("Add property"));
		expect(addBtn.length).toBeGreaterThanOrEqual(1);
		await addBtn[0]?.trigger("click");
		await nextTick();
		// Should have at least one property-value row
		expect(wrapper.findAll(".pv-row").length).toBeGreaterThanOrEqual(1);
	});

	it("should emit created on successful submit", async () => {
		const wrapper = mountWithProviders(CreateIndividualDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const input = wrapper.find(".form-input");
		await input.setValue("JohnDoe");
		const classSelect = wrapper.find(".form-select");
		await classSelect.setValue("owl:Thing");
		const createBtn = wrapper.find(".btn--primary");
		await createBtn.trigger("click");
		await new Promise((resolve) => setTimeout(resolve, 600));
		expect(wrapper.emitted("created")).toBeTruthy();
		expect(wrapper.emitted("created")?.[0]).toEqual(["JohnDoe"]);
	});
});
