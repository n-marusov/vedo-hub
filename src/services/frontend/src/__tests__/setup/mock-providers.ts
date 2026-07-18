// @m2.5 — Mock providers for vitest component tests
// Re-exports from test-utils for plan-specified import path

import { type VueWrapper, mount } from "@vue/test-utils";
import { describe } from "vitest";
import type { ComponentPublicInstance } from "vue";
import { createRouter, createWebHistory } from "vue-router";

// @m2.5 — Creates a mock router with a provided route
export function createMockRouter(initialRoute = "/dashboard/home") {
	const router = createRouter({
		history: createWebHistory(),
		routes: [
			{
				path: "/:pathMatch(.*)*",
				name: "catch-all",
				component: { template: "<div />" },
			},
		],
	});
	router.push(initialRoute);
	return router;
}

// @m2.5 — Mounts a component with common providers (router, stubs)
export function mountWithProviders(
	component: ComponentPublicInstance,
	options: Record<string, unknown> = {},
): VueWrapper {
	const router = createMockRouter();
	return mount(component, {
		global: {
			plugins: [router],
			stubs: {
				"router-link": true,
				"router-view": true,
			},
		},
		...options,
	});
}

// @m2.5 — Wait for async query to settle (flush promises and timers)
export async function waitForQuery(): Promise<void> {
	await new Promise((resolve) => setTimeout(resolve, 50));
}

// @m2.5 — Creates a describePage helper for consistent page test structure
export function describePage(name: string, fn: () => void): void {
	describe(`Page: ${name}`, fn);
}
