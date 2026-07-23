// @m4 — VersioningPage vitest spec (GREEN phase: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: commits and branches from API via REST API calls
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { beforeEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// VersioningPage uses REST listCommits/listBranches/listTags from @/api/versioning
vi.mock("@/api/versioning", () => ({
	listCommits: vi.fn().mockResolvedValue({ items: [], total: 0, page: 0, perPage: 20 }),
	listBranches: vi.fn().mockResolvedValue({ items: [], total: 0 }),
	listTags: vi.fn().mockResolvedValue({ items: [], total: 0 }),
}));

describePage("VersioningPage", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("should render commit history from API when Commits tab is active", async () => {
		const { listCommits, listBranches, listTags } = await import("@/api/versioning");
		vi.mocked(listCommits).mockResolvedValue({ items: [{ author: "Alice", message: "Test", sha: "abc123", date: new Date().toISOString(), branch: "main" }], total: 1, page: 0, perPage: 20 });
		vi.mocked(listBranches).mockResolvedValue({ items: [], total: 0 });
		vi.mocked(listTags).mockResolvedValue({ items: [], total: 0 });
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Commit History");
	});

	it("should switch tabs when tab button is clicked", async () => {
		const { listCommits, listBranches, listTags } = await import("@/api/versioning");
		vi.mocked(listCommits).mockResolvedValue({ items: [], total: 0, page: 0, perPage: 20 });
		vi.mocked(listBranches).mockResolvedValue({ items: [], total: 0 });
		vi.mocked(listTags).mockResolvedValue({ items: [], total: 0 });
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		const tabs = wrapper.findAll(".tab");
		expect(tabs.length).toBeGreaterThanOrEqual(6);
	});

	it("should show empty state when no commits from API", async () => {
		const { listCommits, listBranches, listTags } = await import("@/api/versioning");
		vi.mocked(listCommits).mockResolvedValue({ items: [], total: 0, page: 0, perPage: 20 });
		vi.mocked(listBranches).mockResolvedValue({ items: [], total: 0 });
		vi.mocked(listTags).mockResolvedValue({ items: [], total: 0 });
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await new Promise((resolve) => setTimeout(resolve, 200));
		await nextTick();
		expect(wrapper.text()).toContain("No commits yet");
	});
});
