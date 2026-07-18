// @m2.5 — CreateIndividualDialog vitest spec
// Validates: REQ-USR.UI.gui-implementation
import {
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import CreateIndividualDialog from "@/components/ontology/CreateIndividualDialog.vue";
import { describe, expect, it } from "vitest";
import { nextTick } from "vue";

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
		const createBtn = wrapper.find(".btn--primary");
		await createBtn.trigger("click");
		await new Promise((resolve) => setTimeout(resolve, 600));
		expect(wrapper.emitted("created")).toBeTruthy();
		expect(wrapper.emitted("created")?.[0]).toEqual(["JohnDoe"]);
	});
});
