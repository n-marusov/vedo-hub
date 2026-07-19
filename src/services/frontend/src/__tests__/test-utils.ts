// @m2.5 — Vitest test utilities for M2.5 page component tests
// Re-exports from mock-providers for plan-specified import path

export {
	createMockRouter,
	createMockApolloClient,
	mountWithProviders,
	waitForQuery,
	describePage,
} from "@/__tests__/setup/mock-providers";
