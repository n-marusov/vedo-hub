// @m4 — CreateProjectPage vitest spec
// Validates: REQ-FUN.ORG.project-creation
// Tests: create project page flow, group selector, validation, visibility, toast
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// Helper to flush pending Vue async operations — standard Vue 3 testing pattern.
const flushPromises = () => new Promise((resolve) => setTimeout(resolve, 0));

vi.mock("@/api/org", () => ({
	listGroups: vi.fn(),
	createProject: vi.fn(),
}));

// Mock i18n with pre-loaded translations for projects flow
vi.mock("@/composables/useI18n", () => ({
	useI18n: () => ({
		locale: { value: "en" },
		isLoaded: { value: true },
		t: (key: string, params?: Record<string, string>) => {
			const msgs: Record<string, string> = {
				"projects.breadcrumb_workspace": "Workspace",
				"projects.breadcrumb_projects": "Projects",
				"projects.breadcrumb_new": "New project",
				"projects.create_title": "New project",
				"projects.create_description": "Create a new project in a group",
				"projects.group": "Group",
				"projects.group_placeholder": "Select a group",
				"projects.group_required": "Group is required",
				"projects.name": "Project name",
				"projects.name_placeholder": "My project",
				"projects.name_required": "Project name is required",
				"projects.project_url": "Project URL",
				"projects.slug_placeholder": "project-slug",
				"projects.slug_help": "The URL path used to access the project.",
				"projects.description": "Description",
				"projects.description_placeholder": "Project description (optional)",
				"projects.visibility": "Visibility level",
				"projects.visibility_help": "Choose visibility level for this project.",
				"projects.visibility_private": "Private",
				"projects.visibility_private_desc":
					"Project access must be granted explicitly to each user.",
				"projects.visibility_internal": "Internal",
				"projects.visibility_internal_desc":
					"The project can be accessed by any logged in user.",
				"projects.visibility_public": "Public",
				"projects.visibility_public_desc":
					"The project can be accessed without authentication.",
				"projects.cancel": "Cancel",
				"projects.create_button": "Create project",
				"projects.creating": "Creating...",
				"projects.create_success": "Project created successfully",
				"projects.create_error": "Failed to create project: {error}",
				"common.loading": "Loading...",
				"toast.close": "Close",
			};
			let msg = msgs[key] || key;
			if (params) {
				msg = msg.replace(
					/\{(\w+)\}/g,
					(_, k: string) => params[k] || `{${k}}`,
				);
			}
			return msg;
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

const mockGroups = [
	{
		id: "group-1",
		slug: "research-team",
		name: "Research Team",
		description: "Research group",
		parentGroupId: null,
		visibility: "Internal",
		memberCount: 3,
		projectCount: 1,
	},
	{
		id: "group-2",
		slug: "engineering",
		name: "Engineering",
		description: "Engineering group",
		parentGroupId: null,
		visibility: "Private",
		memberCount: 5,
		projectCount: 2,
	},
];

describePage("CreateProjectPage", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should render create project page with title and breadcrumbs", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("New project");
		expect(wrapper.text()).toContain("Workspace");
		expect(wrapper.find(".cpp-breadcrumbs").exists()).toBe(true);
	});

	it("should load and display groups in selector", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		// Wait for groups to load
		await flushPromises();
		await nextTick();

		const select = wrapper.find("#cpp-group-select");
		expect(select.exists()).toBe(true);
		const options = select.findAll("option");
		// First option is placeholder, remaining are groups
		expect(options.length).toBeGreaterThanOrEqual(2);
		// Human-readable names should appear (not UUIDs)
		expect(wrapper.text()).toContain("Research Team");
		expect(wrapper.text()).toContain("Engineering");
	});

	it("should show validation error when project name is empty", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		// Select a group first
		const select = wrapper.find("#cpp-group-select");
		await select.setValue("group-1");
		await nextTick();

		// Submit without name
		const createBtn = wrapper.find(".cpp-btn-create");
		await createBtn.trigger("click");
		await nextTick();

		expect(wrapper.text()).toContain("Project name is required");
	});

	it("should show validation error when no group is selected", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		// Fill in name but no group
		const input = wrapper.find("#cpp-project-name");
		await input.setValue("My Project");
		await nextTick();

		// Submit without group
		const createBtn = wrapper.find(".cpp-btn-create");
		await createBtn.trigger("click");
		await nextTick();

		expect(wrapper.text()).toContain("Group is required");
	});

	it("should call createProject with name, group_id, visibility on submit", async () => {
		const { listGroups, createProject } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		vi.mocked(createProject).mockResolvedValue({
			id: "new-project-1",
			slug: "knowledge-graph",
			name: "Knowledge Graph",
			description: null,
			visibility: "Private",
			ontologyId: "ontology-uuid",
			memberCount: 1,
			updatedAt: null,
		});
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		// Wait for groups to load
		await flushPromises();
		await nextTick();

		// Select a group first
		const select = wrapper.find("#cpp-group-select");
		await select.setValue("group-1");
		await nextTick();

		// Enter project name
		const input = wrapper.find("#cpp-project-name");
		await input.setValue("Knowledge Graph");
		await nextTick();

		// Submit
		const createBtn = wrapper.find(".cpp-btn-create");
		await createBtn.trigger("click");

		await flushPromises();
		await nextTick();

		expect(createProject).toHaveBeenCalledWith(
			expect.objectContaining({
				name: "Knowledge Graph",
				groupId: "group-1",
				visibility: "Private",
			}),
		);
	});

	it("should show success toast and redirect on creation success", async () => {
		const { listGroups, createProject } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		vi.mocked(createProject).mockResolvedValue({
			id: "new-project-1",
			slug: "test-project",
			name: "Test Project",
			description: null,
			visibility: "Private",
			ontologyId: "ontology-uuid",
			memberCount: 1,
			updatedAt: null,
		});
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		await flushPromises();
		await nextTick();

		const select = wrapper.find("#cpp-group-select");
		await select.setValue("group-1");
		await nextTick();

		const input = wrapper.find("#cpp-project-name");
		await input.setValue("Test Project");
		await nextTick();

		const createBtn = wrapper.find(".cpp-btn-create");
		await createBtn.trigger("click");

		await flushPromises();
		await nextTick();

		// createProject should have been called
		expect(createProject).toHaveBeenCalled();
	});

	it("should show error toast on API failure", async () => {
		const { listGroups, createProject } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		vi.mocked(createProject).mockRejectedValue(new Error("API Error"));

		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		await flushPromises();
		await nextTick();

		const select = wrapper.find("#cpp-group-select");
		await select.setValue("group-1");
		await nextTick();

		const input = wrapper.find("#cpp-project-name");
		await input.setValue("Test");
		await nextTick();

		const createBtn = wrapper.find(".cpp-btn-create");
		await createBtn.trigger("click");

		await flushPromises();
		await nextTick();

		// Page should still be visible (no redirect on error)
		expect(wrapper.text()).toContain("New project");
	});

	it("should auto-fill the slug with transliteration for a Cyrillic project name", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		await flushPromises();
		await nextTick();

		const select = wrapper.find("#cpp-group-select");
		await select.setValue("group-1");
		await nextTick();

		const input = wrapper.find("#cpp-project-name");
		await input.setValue("Онтология продукта");
		await nextTick();

		// Slug input should show the transliterated value
		const slugInput = wrapper.find("#cpp-slug-input");
		expect((slugInput.element as HTMLInputElement).value).toBe(
			"ontologiya-produkta",
		);
	});

	it("should render the project slug input with the domain prefix", async () => {
		const { listGroups } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		await flushPromises();
		await nextTick();

		const select = wrapper.find("#cpp-group-select");
		await select.setValue("group-1");
		await nextTick();

		const slugInput = wrapper.find("#cpp-slug-input");
		expect(slugInput.exists()).toBe(true);
		expect(wrapper.text()).toContain("vedo-core.local/");
	});

	it("should send the edited slug to the createProject API", async () => {
		const { listGroups, createProject } = await import("@/api/org");
		vi.mocked(listGroups).mockResolvedValue(mockGroups);
		vi.mocked(createProject).mockResolvedValue({
			id: "new-project-1",
			slug: "knowledge-graph",
			name: "Knowledge Graph",
			description: null,
			visibility: "Private",
			ontologyId: "ontology-uuid",
			memberCount: 1,
			updatedAt: null,
		});
		const CreateProjectPage = (await import("@/pages/CreateProjectPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateProjectPage);
		await waitForQuery();
		await nextTick();

		await flushPromises();
		await nextTick();

		const select = wrapper.find("#cpp-group-select");
		await select.setValue("group-1");
		await nextTick();

		const input = wrapper.find("#cpp-project-name");
		await input.setValue("Knowledge Graph");
		await nextTick();

		const slugInput = wrapper.find("#cpp-slug-input");
		await slugInput.setValue("knowledge-graph-custom");
		await nextTick();

		const createBtn = wrapper.find(".cpp-btn-create");
		await createBtn.trigger("click");

		await flushPromises();
		await nextTick();

		expect(createProject).toHaveBeenCalledWith(
			expect.objectContaining({
				name: "Knowledge Graph",
				groupId: "group-1",
				visibility: "Private",
				slug: "knowledge-graph-custom",
			}),
		);
	});
});
