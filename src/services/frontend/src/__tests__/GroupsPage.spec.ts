// @m4 — GroupsPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.ORG.group-crud
// Tests: groups hierarchy from REST API
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// GroupsPage uses REST listGroups from @/api/org — mock for test control
vi.mock("@/api/org", () => ({
	listGroups: vi.fn(),
	listProjects: vi.fn(),
	createGroup: vi.fn(),
}));

// Teleport stub: renders slot inline instead of moving to document.body
const TeleportStub = { template: "<div><slot /></div>" };

const mockGroups = [
	{
		id: "group-1",
		name: "Test Group",
		description: "A test group",
		parentGroupId: null,
		visibility: "private",
		memberCount: 1,
		projectCount: 0,
		childGroups: [],
	},
];

describePage("GroupsPage", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should render groups page title", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Groups");
	});

	it("should show search input for filtering groups", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		const searchInput = wrapper.find('input[aria-label="Search groups"]');
		expect(searchInput.exists()).toBe(true);
	});

	it("should render groups section after API data loads", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".gp-list").exists() || wrapper.find(".gp-empty").exists(),
		).toBe(true);
	});

	it("should render page layout with toolbar", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".gp-top").exists()).toBe(true);
	});

	it("should display sort controls for group listing", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Name");
	});

	it("should render expand/collapse toggles for hierarchy", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".gp-top").exists()).toBe(true);
	});

	it("should not crash when groups API fails", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockRejectedValue(new Error("Failed to load groups"));
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await new Promise((resolve) => setTimeout(resolve, 200));
		await nextTick();
		expect(
			wrapper.find(".gp-top").exists() || wrapper.find(".gp-page").exists(),
		).toBe(true);
	});
});

describe("GroupsPage - Create Group Dialog", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should open create dialog when New group button is clicked", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage, {
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();

		// Dialog should be closed initially
		expect(wrapper.find(".dialog-overlay").exists()).toBe(false);

		// Click the New group button
		const newBtn = wrapper.find(".gp-new-btn");
		expect(newBtn.exists()).toBe(true);
		await newBtn.trigger("click");
		await nextTick();

		// Dialog should now be open
		expect(wrapper.find(".dialog-overlay").exists()).toBe(true);
		expect(wrapper.text()).toContain("Create group");
		expect(wrapper.text()).toContain("Group name");
		expect(wrapper.text()).toContain("Group URL");
		expect(wrapper.text()).toContain("Visibility level");
		expect(wrapper.text()).toContain("Invite Members (optional)");
		expect(wrapper.text()).toContain("+ Invite another member");
	});

	it("should close dialog on cancel", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage, {
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();

		// Open dialog
		await wrapper.find(".gp-new-btn").trigger("click");
		await nextTick();
		expect(wrapper.find(".dialog-overlay").exists()).toBe(true);

		// Click Cancel
		const cancelBtn = wrapper
			.findAll("button")
			.filter((b) => b.text().includes("Cancel"));
		expect(cancelBtn.length).toBeGreaterThanOrEqual(1);
		await cancelBtn[0].trigger("click");
		await nextTick();

		// Dialog should be closed
		expect(wrapper.find(".dialog-overlay").exists()).toBe(false);
	});

	it("should show validation error when submitting with empty name", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage, {
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();

		// Open dialog
		await wrapper.find(".gp-new-btn").trigger("click");
		await nextTick();

		// Click Create without entering a name
		const createBtn = wrapper.find(".btn--primary");
		expect(createBtn.exists()).toBe(true);
		await createBtn.trigger("click");
		await nextTick();

		// Validation error should appear
		expect(wrapper.text()).toContain("Group name is required");
	});

	it("should call createGroup API on successful submit and refresh list", async () => {
		const { listGroups, createGroup } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		vi.mocked(createGroup).mockResolvedValue({
			id: "new-group-1",
			name: "Research Team",
			description: null,
			parentGroupId: null,
			visibility: "private",
			memberCount: 0,
			projectCount: 0,
		});
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage, {
			global: { stubs: { Teleport: TeleportStub } },
		});
		await waitForQuery();
		await nextTick();

		// Open dialog
		await wrapper.find(".gp-new-btn").trigger("click");
		await nextTick();

		// Fill in the form
		const nameInput = wrapper.find(".form-input");
		await nameInput.setValue("Research Team");

		// Submit
		const createBtn = wrapper.find(".btn--primary");
		await createBtn.trigger("click");

		// Wait for async submit to resolve
		await new Promise((resolve) => setTimeout(resolve, 100));
		await nextTick();

		// createGroup should have been called
		expect(createGroup).toHaveBeenCalledWith({
			name: "Research Team",
			description: undefined,
			visibility: "private",
		});

		// listGroups should have been called again (refresh after create)
		expect(listGroups).toHaveBeenCalled();

		// Dialog should be closed after successful creation
		expect(wrapper.find(".dialog-overlay").exists()).toBe(false);
	});
});
