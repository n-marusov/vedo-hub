// @m4 — Mock data helpers for Apollo mock link (test/dev only).
//
// After GraphQL tightening: all non-graph mock data (Dashboard, Metrics,
// Deployments, MergeRequests) moved to REST client modules
// (@/api/dashboard.ts, @/api/metrics.ts, etc.). GraphQL mocks remain only
// for graph navigation queries (ClassTree, GraphNeighborhood, etc.).

// @m4 — Helper: create an artificial delay to simulate network latency
export function delay(ms = 200): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}
