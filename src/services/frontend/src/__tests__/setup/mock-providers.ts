// @m2.5 — Mock providers for vitest component tests
// Re-exports from test-utils for plan-specified import path

import {
	MOCK_DASHBOARD_DATA,
	MOCK_DEPLOYMENTS_DATA,
	MOCK_MERGE_REQUESTS_DATA,
	MOCK_METRICS_DATA,
	MOCK_SPARQL_RESULTS,
} from "@/apollo/mock-data";
import type { FetchResult, Operation } from "@apollo/client/core";
import {
	ApolloClient,
	ApolloLink,
	InMemoryCache,
	Observable,
} from "@apollo/client/core";
import { DefaultApolloClient } from "@vue/apollo-composable";
import { type VueWrapper, mount } from "@vue/test-utils";
import { describe } from "vitest";
import type { Component } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";

// @m2.5 — Vitest mock data for operations not covered by mock-data.ts
// Provides realistic defaults so all component tests can mount without a real server
const VITEST_MOCK_RESOLVERS: Record<string, () => unknown> = {
	DashboardAggregate: () => MOCK_DASHBOARD_DATA,
	OntologyMetrics: () => MOCK_METRICS_DATA,
	ListDeployments: () => MOCK_DEPLOYMENTS_DATA,
	ListMergeRequests: () => MOCK_MERGE_REQUESTS_DATA,
	SparqlExecute: () => MOCK_SPARQL_RESULTS,
	RunValidation: () => ({
		runValidation: {
			status: "ok",
			violations: [],
			validatedAt: new Date().toISOString(),
		},
	}),
	// Phase 2-3 queries — provide realistic empty/default data
	GetCommitHistory: () => ({
		commits: { items: [], total: 0, page: 1, perPage: 20 },
	}),
	GetBranches: () => ({
		branches: { items: [], total: 0 },
	}),
	ListProjects: () => ({
		projects: { items: [], total: 0, page: 1, perPage: 20 },
	}),
	ListGroups: () => ({
		groups: [],
	}),
	ListMembers: () => ({
		members: [],
	}),
	GetTags: () => ({
		tags: [],
	}),
	CompareRevisions: () => ({
		compareRevisions: { additions: 0, deletions: 0, changes: [] },
	}),
	UpdateMemberRole: () => ({
		updateMemberRole: {
			success: true,
			member: { id: "1", userId: "u1", role: "editor" },
		},
	}),
	RemoveMember: () => ({
		removeMember: { success: true },
	}),
	Ontology: () => ({
		ontology: {
			id: "test",
			name: "Test",
			branch: "main",
			commit: "abc",
			dirty: false,
		},
	}),
	VersionContext: () => ({
		ontology: { branch: "main", commit: "abc", dirty: false },
	}),
	UpdateDraft: () => ({
		updateDraft: { success: true, timestamp: new Date().toISOString() },
	}),
	GraphNeighborhood: () => ({
		graphNeighborhood: { nodes: [], edges: [] },
	}),
};

// @m2.5 — Apollo link that resolves ALL operations for vitest (no real HTTP)
// Uses known mock resolvers; falls back to empty data/default for unknown operations
class VitestMockLink extends ApolloLink {
	request(operation: Operation): Observable<FetchResult> | null {
		const opName = operation.operationName || "unknown";
		const resolver = VITEST_MOCK_RESOLVERS[opName];

		return new Observable<FetchResult>((observer) => {
			// 20ms delay — faster than the 200ms production mock for quicker tests
			setTimeout(() => {
				if (resolver) {
					observer.next({ data: resolver() as Record<string, unknown> });
				} else {
					// Unknown operation — return empty data to avoid crash
					console.debug(
						JSON.stringify({
							level: "debug",
							msg: "vitest.mock.unhandled_operation",
							operation: opName,
							ts: new Date().toISOString(),
						}),
					);
					observer.next({ data: {} });
				}
				observer.complete();
			}, 20);
		});
	}
}

// @m2.5 — Creates a mock Apollo client for vitest environment
// Components using useQuery/useMutation need an Apollo client via provideApolloClient()
export function createMockApolloClient(): ApolloClient<unknown> {
	return new ApolloClient({
		link: new VitestMockLink() as unknown as ApolloLink,
		cache: new InMemoryCache(),
		defaultOptions: {
			watchQuery: { fetchPolicy: "cache-and-network", errorPolicy: "all" },
			query: { fetchPolicy: "cache-first", errorPolicy: "all" },
		},
	});
}

// @m2.5 — Creates a mock router with a provided route
// Uses createMemoryHistory with initial URL to avoid async navigation issues
export function createMockRouter(initialRoute = "/dashboard/home") {
	const router = createRouter({
		history: createMemoryHistory(initialRoute),
		routes: [
			{
				path: "/:pathMatch(.*)*",
				name: "catch-all",
				component: { template: "<div />" },
			},
		],
	});
	return router;
}

const mockApolloClient = createMockApolloClient();

// @m2.5 — Mounts a component with common providers (router, Apollo client, stubs)
// Deep-merges global options so user-provided stubs/plugins don't override defaults
export function mountWithProviders(
	component: Component,
	options: Record<string, unknown> = {},
): VueWrapper {
	const router = createMockRouter();
	const userGlobal = (options.global as Record<string, unknown>) || {};
	const userStubs = (userGlobal.stubs as Record<string, unknown>) || {};
	const userPlugins = (userGlobal.plugins as unknown[]) || [];
	const userProvide = (userGlobal.provide as Record<string, unknown>) || {};

	return mount(component, {
		...options,
		global: {
			...userGlobal,
			// biome-ignore lint/suspicious/noExplicitAny: Vue Plugin union type mismatch between packages
			plugins: [router, ...(userPlugins as any[])],
			provide: {
				[DefaultApolloClient as symbol]: mockApolloClient,
				...userProvide,
			},
			stubs: {
				"router-link": true,
				"router-view": true,
				...userStubs,
			},
		},
	});
}

// @m2.5 — Wait for async query to settle (flush promises and timers)
export async function waitForQuery(): Promise<void> {
	await new Promise((resolve) => setTimeout(resolve, 100));
}

// @m2.5 — Creates a describePage helper for consistent page test structure
export function describePage(name: string, fn: () => void): void {
	describe(`Page: ${name}`, fn);
}
