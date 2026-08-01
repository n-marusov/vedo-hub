// Apollo Client setup — single GraphQL client for graph navigation queries.
// After GraphQL tightening, only graph navigation queries remain.
// Non-graph reads (versioning, org, comments, metrics, dashboard,
// deployments, merge requests) migrated to REST.

import {
	ApolloClient,
	InMemoryCache,
	createHttpLink,
	from,
} from "@apollo/client/core";
import { setContext } from "@apollo/client/link/context";
import { onError } from "@apollo/client/link/error";
import { RetryLink } from "@apollo/client/link/retry";
import { MockApolloLink, isMockApiEnabled } from "./mock-link";

// structured logging for Apollo operations (observability constraint)
const log = {
	info: (_msg: string, _ctx: Record<string, unknown>) => {},
	error: (msg: string, ctx: Record<string, unknown>) => {
		console.error(
			JSON.stringify({
				level: "error",
				msg,
				...ctx,
				ts: new Date().toISOString(),
			}),
		);
	},
};

const httpLink = createHttpLink({
	uri: window.__VEDO_CONFIG__?.GRAPHQL_ENDPOINT || "/api/v1/graphql",
});

const authLink = setContext((_, { headers }) => {
	const token = localStorage.getItem("vedo-jwt-token");
	return {
		headers: {
			...headers,
			authorization: token ? `Bearer ${token}` : "",
		},
	};
});

const retryLink = new RetryLink({
	delay: {
		initial: 300,
		max: 3000,
		jitter: true,
	},
	attempts: {
		max: 3,
		retryIf: (error, operation) => {
			// Skip retries for mutation operations — they are not idempotent
			if (
				operation.query.definitions.some(
					(def) =>
						def.kind === "OperationDefinition" && def.operation === "mutation",
				)
			) {
				return false;
			}
			return !!error;
		},
	},
});

const errorLink = onError(({ graphQLErrors, networkError, operation }) => {
	if (graphQLErrors) {
		for (const err of graphQLErrors) {
			log.error("apollo.graphql_error", {
				message: err.message,
				locations: err.locations,
				path: err.path,
				operation: operation.operationName,
			});
		}
	}
	if (networkError) {
		log.error("apollo.network_error", {
			message: networkError.message,
			operation: operation.operationName,
		});
	}
});

// When VITE_USE_MOCK_API=true, insert MockApolloLink before retryLink
const mockLink = isMockApiEnabled() ? new MockApolloLink() : null;

const links = mockLink
	? [mockLink, retryLink, errorLink, authLink, httpLink]
	: [retryLink, errorLink, authLink, httpLink];

export const apolloClient = new ApolloClient({
	link: from(links),
	cache: new InMemoryCache({
		typePolicies: {
			Query: {
				fields: {
					ontology: {
						merge(_existing, incoming) {
							return incoming;
						},
					},
				},
			},
		},
	}),
	defaultOptions: {
		watchQuery: {
			fetchPolicy: "cache-and-network",
			errorPolicy: "all",
		},
		query: {
			fetchPolicy: "cache-first",
			errorPolicy: "all",
		},
	},
});

// Clear stale Apollo cache entries on startup.
// After GraphQL tightening, non-graph queries (groups, projects, members,
// commits, dashboard, metrics, deployments, merge_requests, etc.)
// migrated to REST. Cached Apollo entries from previous sessions must
// be evicted so the client does not auto-refetch them.
apolloClient.cache.evict({ fieldName: "groups" });
apolloClient.cache.evict({ fieldName: "projects" });
apolloClient.cache.evict({ fieldName: "members" });
apolloClient.cache.evict({ fieldName: "commits" });
apolloClient.cache.evict({ fieldName: "branch" });
apolloClient.cache.evict({ fieldName: "branches" });
apolloClient.cache.evict({ fieldName: "tags" });
apolloClient.cache.evict({ fieldName: "compareRevisions" });
apolloClient.cache.evict({ fieldName: "dashboard" });
apolloClient.cache.evict({ fieldName: "ontologyMetrics" });
apolloClient.cache.evict({ fieldName: "deployments" });
apolloClient.cache.evict({ fieldName: "mergeRequests" });
apolloClient.cache.evict({ fieldName: "comments" });
apolloClient.cache.evict({ fieldName: "commentFeed" });
apolloClient.cache.evict({ fieldName: "runValidation" });
apolloClient.cache.evict({ fieldName: "userPreferences" });
apolloClient.cache.evict({ fieldName: "ontology" });
apolloClient.cache.gc();

log.info("apollo.client.initialized", {
	endpoint: window.__VEDO_CONFIG__?.GRAPHQL_ENDPOINT || "/api/v1/graphql",
});
