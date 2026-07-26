// REST API client for SPARQL query execution
//
// Per ADR-DES.API.rest-graphql-mutation-boundary.md and
// ADR-DES.API.graphql-sparql-split-strategy.md § «Разделение ответственности»,
// SPARQL execution is only available through the REST endpoint
// `POST /api/v1/sparql` with DoS protection (CircuitBreakerMiddleware, rate
// limiting, query complexity checks). SPARQL via GraphQL (`sparqlQuery`
// resolver) is forbidden because it bypasses the gateway-side DoS defenses.

import axios from 'axios'

const SPARQL_BASE = '/api/v1'

const api = axios.create({
  baseURL: SPARQL_BASE,
  headers: { 'X-Requested-With': 'XMLHttpRequest' }
})

// Inject JWT token from localStorage
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('vedo-jwt-token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export interface SparqlRequest {
  ontologyId: string
  query: string
  limit?: number
  offset?: number
}

// Matches the response shape of REST `POST /api/v1/sparql`
// (handler `queryHandler.HandleSPARQL`). Same column/rows shape that the
// prior GraphQL `sparqlQuery` resolver returned — so the Vue UI can consume
// the data without further transformation.
export interface SparqlResult {
  columns: string[]
  rows: unknown[][]
  total: number
  executionTimeMs: number
}

export interface SparqlErrorResponse {
  error: {
    code: string
    message: string
  }
}

/**
 * Execute a read-only SPARQL query against the ontology through the REST
 * endpoint. Bearer JWT is attached by the request interceptor.
 *
 * Errors:
 *   - 400 SPARQL-SYNTAX-ERROR       — malformed SPARQL
 *   - 400 SPARQL-FORBIDDEN-PATTERN  — pre-filter level 1 blocked the query
 *                                     (see ADR-DES.API.sparql-dos-protection § 2)
 *   - 429 SPARQL-RATE-LIMITED       — gateway rate limit exceeded
 *   - 503 SPARQL-CIRCUIT-OPEN       — circuit breaker is open
 *
 * Re-throws as a plain Error so callers can render `.message` directly.
 */
export async function executeSparql(req: SparqlRequest): Promise<SparqlResult> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'sparql.rest.request',
      ontologyId: req.ontologyId,
      queryLength: req.query?.length ?? 0,
      limit: req.limit,
      offset: req.offset,
      ts: new Date().toISOString()
    })
  )

  try {
    const response = await api.post<SparqlResult>('/sparql', {
      ontologyId: req.ontologyId,
      query: req.query,
      limit: req.limit,
      offset: req.offset
    })

    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'sparql.rest.success',
        ontologyId: req.ontologyId,
        totalResults: response.data.total,
        executionTimeMs: response.data.executionTimeMs,
        ts: new Date().toISOString()
      })
    )

    return response.data
  } catch (err) {
    const axiosErr = err as {
      response?: { data?: SparqlErrorResponse }
      message?: string
    }
    const serverMessage = axiosErr.response?.data?.error?.message
    const message = serverMessage ?? axiosErr.message ?? 'SPARQL execution failed'

    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'sparql.rest.failed',
        ontologyId: req.ontologyId,
        error: message,
        ts: new Date().toISOString()
      })
    )

    throw new Error(message)
  }
}
