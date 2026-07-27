// @m4 — GroupDetailPage vitest spec
// Validates: REQ-FUN.ORG.group-crud
// Tests: group detail rendering, loading, error, visibility badge
import { describePage, waitForQuery } from "@/__tests__/setup/mock-providers";
import { mount } from "@vue/test-utils";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";

vi.mock("@/api/org", () => ({
	getGroup: vi.fn(),
	listGroups: vi.fn(),
}));

vi.mock("@/composables/useI18n", () => ({
	useI18n: () => ({
		locale: { value: "en" },
		isLoaded: { value: true },
		t: (key: string) => {
			const msgs: Record<string, string> = {
				"groups.breadcrumb_workspace": "Workspace",
				"groups.breadcrumb_groups": "Groups",
				"groups.subgroups": "Subgroups",
				"groups.projects": "Projects",
				"groups.members": "Members",
				"groups.retry": "Retry",
				"groups.load_error": "Failed to load groups",
				"common.loading": "Loading...",
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
		showToast: vi.fn(),
		dismissToast: vi.fn(),
	}),
}));

const mockGroup = {
	id: "group-1",
	name: "Engineering",
	description: "Engineering department",
	parentGroupId: null,
	visibility: "private",
	memberCount: 5,
	projectCount: 3,
	childGroups: [],
};

describePage("GroupDetailPage", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	async function mountGroupDetail(id = "group-1") {
		const router = createRouter({
			history: createMemoryHistory(`/dashboard/groups/${id}`),
			routes: [
				{
					path: "/dashboard/groups/:id",
					name: "group-detail",
					component: { template: "<div />" },
				},
			],
		});
		await router.push(`/dashboard/groups/${id}`);
		await router.isReady();

		const GroupDetailPage = (await import("@/pages/GroupDetailPage.vue"))
			.default;
		return mount(GroupDetailPage, {
			global: {
				plugins: [router],
				stubs: { "router-link": true, "router-view": true },
			},
		});
	}

	it("should render group name from API", async () => {
		const { getGroup } = await import("@/api/org");
		vi.mocked(getGroup).mockResolvedValue(mockGroup);
		const wrapper = await mountGroupDetail("group-1");
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Engineering");
	});

	it("should show visibility badge", async () => {
		const { getGroup } = await import("@/api/org");
		vi.mocked(getGroup).mockResolvedValue(mockGroup);
		const wrapper = await mountGroupDetail("group-1");
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".gdp-vis-badge").exists()).toBe(true);
	});

	it("should show member/project/subgroup counts", async () => {
		const { getGroup } = await import("@/api/org");
		vi.mocked(getGroup).mockResolvedValue(mockGroup);
		const wrapper = await mountGroupDetail("group-1");
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("5");
		expect(wrapper.text()).toContain("3");
	});

	it("should show loading state", async () => {
		const { getGroup } = await import("@/api/org");
		vi.mocked(getGroup).mockImplementation(() => new Promise(() => {}));
		const wrapper = await mountGroupDetail("group-1");
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Loading");
	});

	it("should show error state on API failure", async () => {
		const { getGroup } = await import("@/api/org");
		vi.mocked(getGroup).mockRejectedValue(new Error("Not found"));
		const wrapper = await mountGroupDetail("group-1");
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".gdp-state--error").exists()).toBe(true);
	});
});
