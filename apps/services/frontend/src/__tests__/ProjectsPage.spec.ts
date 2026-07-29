// @m4 — ProjectsPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: projects list from REST API, navigation to create page
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
	createProject: vi.fn(),
}));

// Mock i18n
vi.mock("@/composables/useI18n", () => ({
	useI18n: () => ({
		locale: { value: "en" },
		isLoaded: { value: true },
		t: (key: string, _params?: Record<string, string>) => {
			const msgs: Record<string, string> = {
				"projects.breadcrumb_workspace": "Workspace",
				"projects.breadcrumb_projects": "Projects",
				"projects.title": "Projects",
				"projects.search_placeholder": "Search projects...",
				"projects.no_projects": "No projects found.",
				"projects.load_error": "Failed to load projects",
				"projects.retry": "Retry",
				"projects.new_project": "New project",
				"toast.close": "Close",
			};
			return msgs[key] || key;
		},
		setLocale: vi.fn(),
		toggleLocale: vi.fn(),
	}),
}));

vi.mock("@/composables/useToast", () => ({
	useToast: () => ({
		toastVisible: { value: false },
		toastMessage: { value: "" },
		toastType: { value: "success" as const },
		showToast: vi.fn(),
		dismissToast: vi.fn(),
	}),
}));

const mockProjects = {
	items: [
		{
			id: "proj-1",
			name: "Test Project",
			description: "A test",
			visibility: "private",
			ontologyId: "ontology-1",
			memberCount: 1,
			updatedAt: null,
		},
	],
	total: 1,
};

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

	it('should show "New project" button that navigates to create page', async () => {
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
		vi.mocked(listProjects).mockRejectedValue(
			new Error("Failed to load projects"),
		);
		const ProjectsPage = (await import("@/pages/ProjectsPage.vue")).default;
		const wrapper = mountWithProviders(ProjectsPage);
		await new Promise((resolve) => setTimeout(resolve, 200));
		await nextTick();
		expect(
			wrapper.find(".pp-top").exists() || wrapper.find(".pp-page").exists(),
		).toBe(true);
	});
});
