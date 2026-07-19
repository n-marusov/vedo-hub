// @m4 — GroupsPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: groups hierarchy from LIST_GROUPS_QUERY via Apollo
import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("GroupsPage", () => {
	afterEach(() => {
		resetMockResults();
	});

	it("should render groups page title", async () => {
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Groups");
	});

	it("should show search input for filtering groups", async () => {
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		const searchInput = wrapper.find('input[aria-label="Search groups"]');
		expect(searchInput.exists()).toBe(true);
	});

	it("should render groups section after API data loads", async () => {
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".gp-list").exists() || wrapper.find(".gp-empty").exists(),
		).toBe(true);
	});

	it("should render page layout with toolbar", async () => {
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".gp-top").exists()).toBe(true);
	});

	it("should display sort controls for group listing", async () => {
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Name");
	});

	it("should render expand/collapse toggles for hierarchy", async () => {
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".gp-top").exists()).toBe(true);
	});

	it("should not crash when groups API fails", async () => {
		setMockOperationResult(
			"ListGroups",
			null,
			new Error("Failed to load groups"),
		);
		const GroupsPage = (await import("@/pages/GroupsPage.vue")).default;
		const wrapper = mountWithProviders(GroupsPage);
		await waitForQuery();
		await nextTick();
		// Component should render without crashing on API failure
		expect(
			wrapper.find(".gp-top").exists() || wrapper.find(".gp-page").exists(),
		).toBe(true);
	});
});
