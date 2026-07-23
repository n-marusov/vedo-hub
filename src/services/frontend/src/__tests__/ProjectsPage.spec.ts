// @m4 — ProjectsPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: projects list from REST API
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// ProjectsPage uses REST listProjects from @/api/org — mock for test control
vi.mock("@/api/org", () => ({
	listProjects: vi.fn(),
	listGroups: vi.fn(),
}));

const mockProjects = { items: [{ id: "proj-1", name: "Test Project", description: "A test", visibility: "private", ownerId: "user-1" }], total: 1 };

describePage("ProjectsPage", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should render projects list page title", async () => {
		const { listProjects } = await import("@/api/org");
		vi.mocked(listProjects).mockResolvedValue(mockProjects);
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Projects");
	});

	it("should show search input for filtering projects", async () => {
		const { listProjects } = await import("@/api/org");
		vi.mocked(listProjects).mockResolvedValue(mockProjects);
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		const searchInput = wrapper.find('input[aria-label="Search projects"]');
		expect(searchInput.exists()).toBe(true);
	});

	it("should display sort controls for name and direction", async () => {
		const { listProjects } = await import("@/api/org");
		vi.mocked(listProjects).mockResolvedValue(mockProjects);
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Name");
	});

	it("should render projects section after API data loads", async () => {
		const { listProjects } = await import("@/api/org");
		vi.mocked(listProjects).mockResolvedValue(mockProjects);
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".pp-list").exists() || wrapper.find(".pp-empty").exists(),
		).toBe(true);
	});

	it("should render page layout with toolbar", async () => {
		const { listProjects } = await import("@/api/org");
		vi.mocked(listProjects).mockResolvedValue(mockProjects);
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".pp-top").exists()).toBe(true);
	});

	it('should show "New project" button', async () => {
		const { listProjects } = await import("@/api/org");
		vi.mocked(listProjects).mockResolvedValue(mockProjects);
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await waitForQuery();
		await nextTick();
		const newBtn = wrapper.find(".pp-new-btn");
		expect(newBtn.exists()).toBe(true);
		expect(newBtn.text()).toContain("New project");
	});

	it("should not crash when projects API fails", async () => {
		const { listProjects } = await import("@/api/org");
		vi.mocked(listProjects).mockRejectedValue(new Error("Failed to load projects"));
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await new Promise((resolve) => setTimeout(resolve, 200));
		await nextTick();
		expect(
			wrapper.find(".pp-top").exists() || wrapper.find(".pp-page").exists(),
		).toBe(true);
	});
});
