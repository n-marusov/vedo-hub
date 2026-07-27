// @m4 — ORG API vitest spec
// Tests: createGroup, create project, etc. for REQ-USR.GROUPS.create
import { afterEach, describe, expect, it, vi } from "vitest";

// Mock axios before importing the module under test
vi.mock("axios", () => {
	const mockAxiosInstance = {
		get: vi.fn(),
		post: vi.fn(),
		interceptors: {
			request: { use: vi.fn() },
		},
	};
	return {
		default: {
			create: vi.fn(() => mockAxiosInstance),
			isAxiosError: vi.fn(() => true),
		},
	};
});

describe("org API - createGroup", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should POST /groups with name, description and visibility", async () => {
		const axios = await import("axios");
		const mockApi = axios.default.create();
		vi.mocked(mockApi.post).mockResolvedValue({
			data: {
				data: {
					id: "new-group-1",
					name: "Research Team",
					description: "Team working on research",
					visibility: "private",
					parentGroupId: null,
					memberCount: 0,
					projectCount: 0,
				},
			},
		});

		const { createGroup } = await import("@/api/org");
		const result = await createGroup({
			name: "Research Team",
			description: "Team working on research",
			visibility: "private",
		});

		expect(mockApi.post).toHaveBeenCalledWith(
			"/groups",
			expect.objectContaining({
				name: "Research Team",
				description: "Team working on research",
				visibility: "private",
			}),
		);
		expect(result).toEqual({
			id: "new-group-1",
			name: "Research Team",
			description: "Team working on research",
			visibility: "private",
			parentGroupId: null,
			memberCount: 0,
			projectCount: 0,
		});
	});

	it("should POST /groups with minimal data (name only)", async () => {
		const axios = await import("axios");
		const mockApi = axios.default.create();
		vi.mocked(mockApi.post).mockResolvedValue({
			data: {
				data: {
					id: "new-group-2",
					name: "Minimal Group",
					description: "",
					visibility: "private",
					parentGroupId: null,
					memberCount: 0,
					projectCount: 0,
				},
			},
		});

		const { createGroup } = await import("@/api/org");
		const result = await createGroup({
			name: "Minimal Group",
		});

		expect(mockApi.post).toHaveBeenCalledWith(
			"/groups",
			expect.objectContaining({
				name: "Minimal Group",
			}),
		);
		expect(result.name).toBe("Minimal Group");
	});

	it("should include parent_id in createGroup payload", async () => {
		const axios = await import("axios");
		const mockApi = axios.default.create();
		vi.mocked(mockApi.post).mockResolvedValue({
			data: {
				data: {
					id: "new-subgroup-1",
					name: "Frontend",
					parentGroupId: "parent-1",
					visibility: "private",
					memberCount: 0,
					projectCount: 0,
				},
			},
		});

		const { createGroup } = await import("@/api/org");
		await createGroup({
			name: "Frontend",
			parent_id: "parent-1",
		});

		expect(mockApi.post).toHaveBeenCalledWith(
			"/groups",
			expect.objectContaining({
				name: "Frontend",
				parent_id: "parent-1",
			}),
		);
	});

	it("should include slug in createGroup payload when provided", async () => {
		const axios = await import("axios");
		const mockApi = axios.default.create();
		vi.mocked(mockApi.post).mockResolvedValue({
			data: {
				data: {
					id: "new-group-3",
					name: "My Test Group",
					parentGroupId: null,
					visibility: "private",
					memberCount: 0,
					projectCount: 0,
				},
			},
		});

		const { createGroup } = await import("@/api/org");
		await createGroup({
			name: "My Test Group",
			slug: "my-test-group",
		});

		expect(mockApi.post).toHaveBeenCalledWith(
			"/groups",
			expect.objectContaining({
				name: "My Test Group",
				slug: "my-test-group",
			}),
		);
	});

	it("should throw error when API fails", async () => {
		const axios = await import("axios");
		const mockApi = axios.default.create();
		vi.mocked(mockApi.post).mockRejectedValue({
			response: {
				data: {
					error: { message: "Group name already exists" },
				},
			},
		});

		const { createGroup } = await import("@/api/org");
		await expect(createGroup({ name: "Duplicate" })).rejects.toThrow(
			"Group name already exists",
		);
	});

	it("should throw generic error when no response body", async () => {
		const axios = await import("axios");
		const mockApi = axios.default.create();
		vi.mocked(mockApi.post).mockRejectedValue(new Error("Network Error"));

		const { createGroup } = await import("@/api/org");
		await expect(createGroup({ name: "Failing" })).rejects.toThrow(
			"Network Error",
		);
	});
});
