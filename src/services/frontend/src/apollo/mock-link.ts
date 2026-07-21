// @m4 — Apollo mock link for Block В (Backend Pages)
// Intercepts specific queries with realistic mock data when VITE_USE_MOCK_API=true

import type { FetchResult, Operation } from "@apollo/client/core";
import { ApolloLink, Observable } from "@apollo/client/core";
import {
	MOCK_DASHBOARD_DATA,
	MOCK_DEPLOYMENTS_DATA,
	MOCK_MERGE_REQUESTS_DATA,
	MOCK_METRICS_DATA,
	delay,
} from "./mock-data";

// Map from query name to mock data resolver
// NOTE: SPARQL is REST-only per ADR-DES.API.rest-graphql-mutation-boundary.md
// — `SparqlExecute` GraphQL resolver removed; use `POST /api/v1/sparql`.
const mockResolvers: Record<string, () => unknown> = {
	DashboardAggregate: () => MOCK_DASHBOARD_DATA,
	OntologyMetrics: () => MOCK_METRICS_DATA,
	ListDeployments: () => MOCK_DEPLOYMENTS_DATA,
	ListMergeRequests: () => MOCK_MERGE_REQUESTS_DATA,
	RunValidation: () => ({
		runValidation: {
			status: "ok",
			violations: [],
			validatedAt: new Date().toISOString(),
		},
	}),
};

export class MockApolloLink extends ApolloLink {
	request(operation: Operation): Observable<FetchResult> | null {
		const opName = operation.operationName || "unknown";

		// Check if we have mock data for this operation
		const resolver = mockResolvers[opName];
		if (!resolver) {
			return null; // forward to next link for unmocked operations
		}

		// Log mock usage
		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "mock.link.intercepted",
				operation: opName,
				ts: new Date().toISOString(),
			}),
		);

		// Apply artificial delay for realistic behavior
		return new Observable<FetchResult>((observer) => {
			delay(200).then(() => {
				observer.next({ data: resolver() as Record<string, unknown> });
				observer.complete();
			});
		});
	}
}

// @m4 — Check whether mock API mode is enabled
export function isMockApiEnabled(): boolean {
	return import.meta.env.VITE_USE_MOCK_API === "true";
}
