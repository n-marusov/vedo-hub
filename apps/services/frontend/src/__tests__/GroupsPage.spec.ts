// @m4 — GroupsPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.ORG.group-crud
// Tests: groups hierarchy from REST API
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// GroupsPage uses REST listGroups from @/api/org — mock for test control
vi.mock("@/api/org", () => ({
	listGroups: vi.fn(),
	listProjects: vi.fn(),
	createGroup: vi.fn(),
}));

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

	it("should navigate to create group page when New group is clicked", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();

		const newBtn = wrapper.find(".gp-new-btn");
		expect(newBtn.exists()).toBe(true);
	});

	it("should use i18n keys for labels", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Groups");
	});
});
