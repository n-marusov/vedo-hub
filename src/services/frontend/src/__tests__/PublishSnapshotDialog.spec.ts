// @m2.5 — PublishSnapshotDialog vitest spec
// Validates: REQ-USR.UI.gui-implementation
import {
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import PublishSnapshotDialog from "@/components/ontology/PublishSnapshotDialog.vue";
import { describe, expect, it } from "vitest";
import { nextTick } from "vue";

const TeleportStub = { template: "<div><slot /></div>" };

describe("PublishSnapshotDialog", () => {
	it("should render when open is true", async () => {
		const wrapper = mountWithProviders(PublishSnapshotDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dialog-overlay").exists()).toBe(true);
		expect(wrapper.text()).toContain("Publish Snapshot");
	});

	it("should show visibility select and version input", async () => {
		const wrapper = mountWithProviders(PublishSnapshotDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		expect(wrapper.find(".form-select").exists()).toBe(true);
		expect(wrapper.findAll(".form-input").length).toBeGreaterThanOrEqual(1);
	});

	it("should disable Publish button when version is empty", async () => {
		const wrapper = mountWithProviders(PublishSnapshotDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const publishBtn = wrapper.find(".btn--primary");
		expect(publishBtn.attributes("disabled")).toBeDefined();
	});

	it("should show generated URL preview when version is entered", async () => {
		const wrapper = mountWithProviders(PublishSnapshotDialog, {
			props: { open: true, ontologyId: "test-onto" },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const inputs = wrapper.findAll(".form-input");
		await inputs[0]?.setValue("v1.0.0");
		await nextTick();
		// With version set, publish button should be enabled
		const publishBtn = wrapper.find(".btn--primary");
		expect(publishBtn.attributes("disabled")).toBeUndefined();
	});

	it("should emit published on successful publish", async () => {
		const wrapper = mountWithProviders(PublishSnapshotDialog, {
			props: { open: true },
			global: { stubs: { Teleport: TeleportStub } },
		});
		await nextTick();
		const inputs = wrapper.findAll(".form-input");
		await inputs[0]?.setValue("v1.0.0");
		await nextTick();
		const publishBtn = wrapper.find(".btn--primary");
		await publishBtn.trigger("click");
		await new Promise((resolve) => setTimeout(resolve, 1000));
		expect(wrapper.emitted("published")).toBeTruthy();
		expect(wrapper.emitted("published")?.[0]).toEqual(["v1.0.0"]);
	});
});
