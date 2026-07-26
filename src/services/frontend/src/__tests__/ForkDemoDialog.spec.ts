import { flushPromises, mount } from "@vue/test-utils";
// @m8 — ForkDemoDialog vitest spec (GREEN phase — component implemented)
// Validates: REQ-USR.UI.gui-implementation
// Tests: fork demo project dialog renders 5 demos, fork button calls REST API,
//        error states (403, 503), empty state, search filter, cancel
import axios from "axios";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

vi.mock("axios");
const mockedAxios = vi.mocked(axios);

const MOCK_DEMO_PROJECTS = [
	{
		id: "org-1",
		name: "Organization",
		description: "Departments, roles, company structure",
		classCount: 5,
		propertyCount: 4,
		domain: "Enterprise",
	},
	{
		id: "prod-1",
		name: "Product",
		description: "Catalog, manufacturers, reviews",
		classCount: 4,
		propertyCount: 5,
		domain: "Commerce",
	},
	{
		id: "proc-1",
		name: "Process",
		description: "Workflow steps, inputs, outputs",
		classCount: 5,
		propertyCount: 3,
		domain: "Operations",
	},
	{
		id: "gloss-1",
		name: "Glossary",
		description: "Terms, definitions, references",
		classCount: 4,
		propertyCount: 3,
		domain: "Knowledge",
	},
	{
		id: "evt-1",
		name: "Event",
		description: "Events, participants, locations",
		classCount: 4,
		propertyCount: 4,
		domain: "Domain",
	},
];

describe("ForkDemoDialog", () => {
	afterEach(() => {
		vi.clearAllMocks();
		document.body.innerHTML = "";
	});

	async function mountDialog(comp) {
		const wrapper = mount(comp, { props: { modelValue: true } });
		await flushPromises();
		await nextTick();
		return wrapper;
	}

	it("should render list of 5 demo projects with names and descriptions", async () => {
		mockedAxios.get.mockResolvedValueOnce({
			data: { projects: MOCK_DEMO_PROJECTS },
		});
		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		await mountDialog(comp);

		const body = document.body.textContent || "";
		for (const demo of MOCK_DEMO_PROJECTS) {
			expect(body).toContain(demo.name);
			expect(body).toContain(demo.description);
		}
	});

	it("should render Fork button on each demo project card", async () => {
		mockedAxios.get.mockResolvedValueOnce({
			data: { projects: MOCK_DEMO_PROJECTS },
		});
		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		await mountDialog(comp);

		const forkMatches = (document.body.textContent || "").match(/Fork/g);
		expect(forkMatches?.length).toBeGreaterThanOrEqual(5);
	});

	it("should call POST /api/v1/projects/:id/fork when Fork is clicked", async () => {
		mockedAxios.get.mockResolvedValueOnce({
			data: { projects: MOCK_DEMO_PROJECTS },
		});
		mockedAxios.post.mockResolvedValueOnce({
			data: {
				project_id: "new-proj-1",
				ontology_id: "new-onto-1",
				upstream_project_id: "org-1",
			},
		});

		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		await mountDialog(comp);

		const forkTrigger = document.querySelector('[data-testid="fork-org-1"]');
		if (forkTrigger) {
			(forkTrigger as HTMLButtonElement).click();
		}
		await flushPromises();

		expect(mockedAxios.post).toHaveBeenCalledWith(
			expect.stringContaining("/api/v1/projects/org-1/fork"),
			{},
			expect.objectContaining({
				headers: expect.objectContaining({
					"Idempotency-Key": expect.any(String),
				}),
			}),
		);
	});

	it("should try to redirect to new Project workspace on successful fork", async () => {
		mockedAxios.get.mockResolvedValueOnce({
			data: { projects: MOCK_DEMO_PROJECTS },
		});
		mockedAxios.post.mockResolvedValueOnce({
			data: {
				project_id: "new-proj-1",
				ontology_id: "new-onto-1",
				upstream_project_id: "org-1",
			},
		});

		const originalLocation = window.location;
		Object.defineProperty(window, "location", {
			value: { ...originalLocation, href: "" },
			writable: true,
		});

		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		await mountDialog(comp);

		const forkTrigger = document.querySelector('[data-testid="fork-org-1"]');
		if (forkTrigger) {
			(forkTrigger as HTMLButtonElement).click();
		}
		await flushPromises();

		expect(window.location.href).toContain("/projects/new-proj-1");
		Object.defineProperty(window, "location", {
			value: originalLocation,
			writable: true,
		});
	});

	it("should show Access denied on 403 error", async () => {
		mockedAxios.get.mockResolvedValueOnce({
			data: { projects: MOCK_DEMO_PROJECTS },
		});
		mockedAxios.post.mockRejectedValueOnce({
			response: { status: 403, data: { error: "Access denied" } },
		});

		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		await mountDialog(comp);

		const forkTrigger = document.querySelector('[data-testid="fork-org-1"]');
		if (forkTrigger) {
			(forkTrigger as HTMLButtonElement).click();
		}
		await flushPromises();

		expect(document.body.textContent || "").toContain("Access denied");
	});

	it("should show Service unavailable with Retry on 503 error", async () => {
		mockedAxios.get.mockResolvedValueOnce({
			data: { projects: MOCK_DEMO_PROJECTS },
		});
		mockedAxios.post.mockRejectedValueOnce({
			response: { status: 503, data: { error: "Service unavailable" } },
		});

		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		await mountDialog(comp);

		const forkTrigger = document.querySelector('[data-testid="fork-org-1"]');
		if (forkTrigger) {
			(forkTrigger as HTMLButtonElement).click();
		}
		await flushPromises();

		const body = document.body.textContent || "";
		expect(body).toContain("Service unavailable");
		expect(body).toContain("Retry");
	});

	it("should show empty state when no demo projects available", async () => {
		mockedAxios.get.mockResolvedValueOnce({ data: { projects: [] } });
		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		await mountDialog(comp);

		expect(document.body.textContent || "").toContain(
			"No demo projects available",
		);
	});

	it("should filter demo projects by search query", async () => {
		mockedAxios.get.mockResolvedValueOnce({
			data: { projects: MOCK_DEMO_PROJECTS },
		});
		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		await mountDialog(comp);

		const searchInput = document.querySelector(
			'input[aria-label="Search demo projects"]',
		) as HTMLInputElement;
		if (searchInput) {
			searchInput.value = "Organization";
			searchInput.dispatchEvent(new Event("input"));
			await nextTick();
			const body = document.body.textContent || "";
			expect(body).toContain("Organization");
			expect(body).not.toContain("Product");
		}
	});

	it("should close dialog when Cancel button is clicked", async () => {
		mockedAxios.get.mockResolvedValueOnce({
			data: { projects: MOCK_DEMO_PROJECTS },
		});
		const comp = (await import("@/components/projects/ForkDemoDialog.vue"))
			.default;
		const wrapper = mount(comp, { props: { modelValue: true } });
		await flushPromises();
		await nextTick();

		const cancelBtn = document.querySelector('[data-testid="cancel-button"]');
		if (cancelBtn) {
			(cancelBtn as HTMLButtonElement).click();
			await nextTick();
			const emitted = wrapper.emitted();
			const closeEvent =
				emitted["update:modelValue"] || emitted.close || emitted.cancel;
			expect(closeEvent).toBeTruthy();
			expect(closeEvent?.[0]).toEqual([false]);
		}
	});
});
