// @m4 — Mock providers for vitest component tests.
// After GraphQL tightening: non-graph mock resolvers removed (migrated to
// REST clients). Graph navigation mock resolvers remain (ClassTree,
// GraphNeighborhood).

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
import type { Component, Plugin } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";

// Per-test result overrides for edge-case simulation.
const overrideResults: Map<
	string,
	() => {
		data?: Record<string, unknown>;
		error?: Error;
		graphQLErrors?: Array<{ message: string }>;
	}
> = new Map();

export function setMockOperationResult(
	operationName: string,
	data: Record<string, unknown> | null,
	error?: Error,
	graphQLErrors?: Array<{ message: string }>,
): void {
	if (error) {
		overrideResults.set(operationName, () => ({ error }));
	} else if (graphQLErrors) {
		overrideResults.set(operationName, () => ({
			data: data || {},
			graphQLErrors,
		}));
	} else if (data) {
		overrideResults.set(operationName, () => ({ data }));
	} else {
		overrideResults.delete(operationName);
	}
}

export function resetMockResults(): void {
	overrideResults.clear();
}

// Graph-only mock resolvers retained after GraphQL tightening.
// All non-graph operations (versioning, org, comments, metrics, dashboard,
// deployments, merge_requests, validation, draft) migrated to REST clients.
const VITEST_MOCK_RESOLVERS: Record<string, () => unknown> = {
	// Class queries (graph navigation)
	ClassTree: () => ({
		classTree: [
			{
				id: "cls-1",
				label: "owl:Thing",
				children: [{ id: "cls-2", label: "Person", children: [] }],
			},
		],
	}),
	// Graph neighborhood
	GraphNeighborhood: () => ({
		graphNeighborhood: { nodes: [], edges: [] },
	}),
};

// Apollo link that resolves graph navigation operations for vitest.
class VitestMockLink extends ApolloLink {
	request(operation: Operation): Observable<FetchResult> | null {
		const opName = operation.operationName || "unknown";
		const resolver = VITEST_MOCK_RESOLVERS[opName];
		const override = overrideResults.get(opName);

		return new Observable<FetchResult>((observer) => {
			setTimeout(() => {
				if (override) {
					const result = override();
					if (result.error) {
						observer.error(result.error);
					} else if (result.graphQLErrors) {
						observer.next({
							data: result.data as Record<string, unknown>,
							errors: result.graphQLErrors,
						});
						observer.complete();
					} else if (result.data) {
						observer.next({ data: result.data });
						observer.complete();
					}
					return;
				}
				if (resolver) {
					observer.next({ data: resolver() as Record<string, unknown> });
				} else {
					// Unknown operation — return empty data
					observer.next({ data: {} });
				}
				observer.complete();
			}, 20);
		});
	}
}

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

export function mountWithProviders(
	component: Component,
	options: Record<string, unknown> = {},
): VueWrapper {
	const router = createMockRouter();
	const apolloClient = createMockApolloClient();
	const userGlobal = (options.global as Record<string, unknown>) || {};
	const userStubs = (userGlobal.stubs as Record<string, unknown>) || {};
	const userPlugins = (userGlobal.plugins as unknown[]) || [];
	const userProvide = (userGlobal.provide as Record<string, unknown>) || {};

	return mount(component, {
		...options,
		global: {
			...userGlobal,
			plugins: [router, ...userPlugins] as Plugin[],
			provide: {
				[DefaultApolloClient as symbol]: apolloClient,
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

export async function waitForQuery(): Promise<void> {
	await new Promise((resolve) => setTimeout(resolve, 100));
}

export function describePage(name: string, fn: () => void): void {
	describe(`Page: ${name}`, fn);
}
