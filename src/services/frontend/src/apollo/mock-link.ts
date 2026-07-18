// @m2.5 — Apollo mock link for Block В (Backend Pages)
// Intercepts specific queries with realistic mock data when VITE_USE_MOCK_API=true

import type { Operation } from "@apollo/client/core";
import { ApolloLink, Observable } from "@apollo/client/core";
import { print } from "graphql";
import {
	MOCK_DASHBOARD_DATA,
	MOCK_DEPLOYMENTS_DATA,
	MOCK_MERGE_REQUESTS_DATA,
	MOCK_METRICS_DATA,
	delay,
} from "./mock-data";

// Map from query name to mock data resolver
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
	request(operation: Operation): Observable<unknown> {
		return new Observable((observer) => {
			const definition = operation.query.definitions.find(
				(d) => d.kind === "OperationDefinition",
			);
			// Extract operation name from the printed query or the operation context
			const opName = operation.operationName || "unknown";

			// Log mock usage
			console.debug(
				JSON.stringify({
					level: "debug",
					msg: "mock.link.intercepted",
					operation: opName,
					ts: new Date().toISOString(),
				}),
			);

			// Check if we have mock data for this operation
			const resolver = mockResolvers[opName];
			if (!resolver) {
				// Forward to next link for unmocked operations
				observer.complete();
				return;
			}

			// Apply artificial delay for realistic behavior
			delay(200).then(() => {
				const data = resolver();
				observer.next({ data });
				observer.complete();
			});
		});
	}
}

// @m2.5 — Check whether mock API mode is enabled
export function isMockApiEnabled(): boolean {
	return import.meta.env.VITE_USE_MOCK_API === "true";
}
