// @m4 — Apollo mock link for development (VITE_USE_MOCK_API=true).
//
// After GraphQL tightening: all non-graph mock resolvers (Dashboard,
// Metrics, Deployments, MergeRequests, Validation) migrated to REST clients.
// This mock link now passes through all operations to the real backend
// since only graph navigation queries remain in the GraphQL schema.

import type { FetchResult, Operation } from "@apollo/client/core";
import { ApolloLink, type Observable } from "@apollo/client/core";

export class MockApolloLink extends ApolloLink {
	// All previously mocked non-graph operations (DashboardAggregate,
	// OntologyMetrics, ListDeployments, ListMergeRequests, RunValidation)
	// are now REST-only. The mock link passes through to the real backend.
	request(_operation: Operation): Observable<FetchResult> | null {
		return null; // forward to next link
	}
}

// @m4 — Check whether mock API mode is enabled
export function isMockApiEnabled(): boolean {
	return (window.__VEDO_CONFIG__?.USE_MOCK_API || "false") === "true";
}
