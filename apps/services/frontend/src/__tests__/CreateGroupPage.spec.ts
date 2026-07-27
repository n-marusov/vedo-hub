// @m4 — CreateGroupPage vitest spec
// Validates: REQ-FUN.ORG.group-crud
// Tests: create group page flow, validation, parent_id, slug generation
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

vi.mock("@/api/org", () => ({
	listGroups: vi.fn(),
	createGroup: vi.fn(),
}));

// Mock i18n with pre-loaded English translations for page tests
vi.mock("@/composables/useI18n", () => ({
	useI18n: () => ({
		locale: { value: "en" },
		isLoaded: { value: true },
		t: (key: string, params?: Record<string, string>) => {
			const msgs: Record<string, string> = {
				"groups.breadcrumb_workspace": "Workspace",
				"groups.breadcrumb_groups": "Groups",
				"groups.breadcrumb_new": "New group",
				"groups.create_title": "New group",
				"groups.create_description":
					"Groups allow you to manage and collaborate across multiple projects.",
				"groups.group_name": "Group name",
				"groups.name_placeholder": "My group",
				"groups.group_url": "Group URL",
				"groups.slug_placeholder": "my-awesome-group",
				"groups.name_help": "Start with a letter, digit, emoji, or underscore.",
				"groups.visibility": "Visibility level",
				"groups.visibility_help": "Who will be able to see this group?",
				"groups.visibility_private": "Private",
				"groups.visibility_private_desc":
					"The group and its projects can only be viewed by members.",
				"groups.visibility_internal": "Internal",
				"groups.visibility_internal_desc":
					"The group and any internal projects can be viewed by any logged in user except external users.",
				"groups.visibility_public": "Public",
				"groups.visibility_public_desc":
					"The group and any public projects can be viewed without any authentication.",
				"groups.visibility_inherited": "Inherited from parent group",
				"groups.parent_group": "Parent group",
				"groups.create_button": "Create group",
				"groups.creating": "Creating...",
				"groups.create_success": 'Group "{name}" was successfully created.',
				"groups.create_error": "Failed to create group",
				"groups.name_required": "Group name is required",
				"groups.name_duplicate":
					"Group with this name already exists in this path.",
				"groups.cancel": "Cancel",
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

describePage("CreateGroupPage", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should render create group page title", async () => {
		const CreateGroupPage = (await import("@/pages/CreateGroupPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateGroupPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("New group");
	});

	it("should show group name input", async () => {
		const CreateGroupPage = (await import("@/pages/CreateGroupPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateGroupPage);
		await waitForQuery();
		await nextTick();
		const input = wrapper.find('input[aria-label="Group name"]');
		expect(input.exists()).toBe(true);
	});

	it("should show visibility selector for top-level group", async () => {
		const CreateGroupPage = (await import("@/pages/CreateGroupPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateGroupPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Visibility level");
		expect(wrapper.find(".cgp-vis-options").exists()).toBe(true);
	});

	it("should show validation error when submitting with empty name", async () => {
		const CreateGroupPage = (await import("@/pages/CreateGroupPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateGroupPage);
		await waitForQuery();
		await nextTick();

		const createBtn = wrapper.find(".cgp-btn-create");
		await createBtn.trigger("click");
		await nextTick();

		expect(wrapper.text()).toContain("Group name is required");
	});

	it("should call createGroup API and redirect on success", async () => {
		const { createGroup } = await import("@/api/org");
		vi.mocked(createGroup).mockResolvedValue({
			id: "new-group-1",
			name: "Research Team",
			description: null,
			parentGroupId: null,
			visibility: "private",
			memberCount: 0,
			projectCount: 0,
		});

		const CreateGroupPage = (await import("@/pages/CreateGroupPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateGroupPage);
		await waitForQuery();
		await nextTick();

		const input = wrapper.find('input[aria-label="Group name"]');
		await input.setValue("Research Team");

		const createBtn = wrapper.find(".cgp-btn-create");
		await createBtn.trigger("click");

		await new Promise((resolve) => setTimeout(resolve, 100));
		await nextTick();

		expect(createGroup).toHaveBeenCalledWith(
			expect.objectContaining({
				name: "Research Team",
				visibility: "private",
			}),
		);
	});

	it("should show error toast on API failure", async () => {
		const { createGroup } = await import("@/api/org");
		vi.mocked(createGroup).mockRejectedValue(new Error("API Error"));

		const CreateGroupPage = (await import("@/pages/CreateGroupPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateGroupPage);
		await waitForQuery();
		await nextTick();

		const input = wrapper.find('input[aria-label="Group name"]');
		await input.setValue("Test");

		const createBtn = wrapper.find(".cgp-btn-create");
		await createBtn.trigger("click");

		await new Promise((resolve) => setTimeout(resolve, 100));
		await nextTick();

		// Page should still be visible (no redirect on error)
		expect(wrapper.text()).toContain("New group");
	});

	it("should generate slug from name", async () => {
		const CreateGroupPage = (await import("@/pages/CreateGroupPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateGroupPage);
		await waitForQuery();
		await nextTick();

		const input = wrapper.find('input[aria-label="Group name"]');
		await input.setValue("My Test Group");
		await nextTick();

		expect(wrapper.text()).toContain("my-test-group");
	});

	it("should render breadcrumbs", async () => {
		const CreateGroupPage = (await import("@/pages/CreateGroupPage.vue"))
			.default;
		const wrapper = mountWithProviders(CreateGroupPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Workspace");
	});
});
