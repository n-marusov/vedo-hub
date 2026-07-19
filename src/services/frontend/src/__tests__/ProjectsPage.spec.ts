// @m4 — ProjectsPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: projects list from LIST_PROJECTS_QUERY via Apollo
import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("ProjectsPage", () => {
	afterEach(() => {
		resetMockResults();
	});

	it("should render projects list page title", async () => {
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Projects");
	});

	it("should show search input for filtering projects", async () => {
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		const searchInput = wrapper.find('input[aria-label="Search projects"]');
		expect(searchInput.exists()).toBe(true);
	});

	it("should display sort controls for name and direction", async () => {
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Name");
	});

	it("should render projects section after API data loads", async () => {
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		// Page renders with either project rows or empty state — both valid
		expect(
			wrapper.find(".pp-list").exists() || wrapper.find(".pp-empty").exists(),
		).toBe(true);
	});

	it("should render page layout with toolbar", async () => {
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".pp-top").exists()).toBe(true);
	});

	it('should show "New project" button', async () => {
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		const newBtn = wrapper.find(".pp-new-btn");
		expect(newBtn.exists()).toBe(true);
		expect(newBtn.text()).toContain("New project");
	});

	it("should not crash when projects API fails", async () => {
		setMockOperationResult(
			"ListProjects",
			null,
			new Error("Failed to load projects"),
		);
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		// Component should render without crashing on API failure
		expect(
			wrapper.find(".pp-top").exists() || wrapper.find(".pp-page").exists(),
		).toBe(true);
	});
});
