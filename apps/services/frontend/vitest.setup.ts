// @m2.5 — Vitest global setup: configures Apollo client for all component tests
// Components using useQuery/useMutation need Apollo client provided at the app level.
// This setup ensures all mount() calls get the Apollo client automatically.

// NOTE: `MOCK_SPARQL_RESULTS` removed — SPARQL is REST-only per
// ADR-DES.API.rest-graphql-mutation-boundary.md. Use `@/api/sparql` (axios).
import {
  MOCK_DASHBOARD_DATA,
  MOCK_DEPLOYMENTS_DATA,
  MOCK_MERGE_REQUESTS_DATA,
  MOCK_METRICS_DATA
} from '@/apollo/mock-data'
import { ApolloClient, ApolloLink, InMemoryCache } from '@apollo/client/core'
import type { FetchResult, Operation } from '@apollo/client/core'
import { Observable } from '@apollo/client/core'
import { DefaultApolloClient } from '@vue/apollo-composable'
import { config } from '@vue/test-utils'

const VITEST_MOCK_RESOLVERS: Record<string, () => unknown> = {
  DashboardAggregate: () => MOCK_DASHBOARD_DATA,
  OntologyMetrics: () => MOCK_METRICS_DATA,
  ListDeployments: () => MOCK_DEPLOYMENTS_DATA,
  ListMergeRequests: () => MOCK_MERGE_REQUESTS_DATA,
  RunValidation: () => ({
    runValidation: {
      status: 'ok',
      violations: [],
      validatedAt: new Date().toISOString()
    }
  }),
  GetCommitHistory: () => ({
    commits: { items: [], total: 0, page: 1, perPage: 20 }
  }),
  GetBranches: () => ({ branches: { items: [], total: 0 } }),
  ListProjects: () => ({
    projects: { items: [], total: 0, page: 1, perPage: 20 }
  }),
  ListGroups: () => ({ groups: [] }),
  ListMembers: () => ({ members: [] }),
  GetTags: () => ({ tags: [] }),
  CompareRevisions: () => ({
    compareRevisions: { additions: 0, deletions: 0, changes: [] }
  }),
  UpdateMemberRole: () => ({
    updateMemberRole: {
      success: true,
      member: { id: '1', userId: 'u1', role: 'editor' }
    }
  }),
  RemoveMember: () => ({ removeMember: { success: true } }),
  Ontology: () => ({
    ontology: {
      id: 'test',
      name: 'Test',
      branch: 'main',
      commit: 'abc',
      dirty: false
    }
  }),
  VersionContext: () => ({
    ontology: { branch: 'main', commit: 'abc', dirty: false }
  }),
  UpdateDraft: () => ({
    updateDraft: { success: true, timestamp: new Date().toISOString() }
  }),
  GraphNeighborhood: () => ({ graphNeighborhood: { nodes: [], edges: [] } })
}

class VitestMockLink extends ApolloLink {
  request(operation: Operation): Observable<FetchResult> | null {
    const opName = operation.operationName || 'unknown'
    const resolver = VITEST_MOCK_RESOLVERS[opName]
    return new Observable<FetchResult>((observer) => {
      setTimeout(() => {
        observer.next({
          data: resolver ? (resolver() as Record<string, unknown>) : {}
        })
        observer.complete()
      }, 10) // Fast for tests
    })
  }
}

const mockClient = new ApolloClient({
  link: new VitestMockLink() as unknown as ApolloLink,
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: { fetchPolicy: 'cache-and-network', errorPolicy: 'all' },
    query: { fetchPolicy: 'cache-first', errorPolicy: 'all' }
  }
})

// @m2.5 — Auto-provide Apollo client in ALL test mounts
config.global.provide = {
  ...config.global.provide,
  [DefaultApolloClient as symbol]: mockClient
}
