// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Apollo Client setup — single GraphQL client for all frontend data operations
// @hlv:sec [INPUT_VALIDATION] — no secrets in Apollo Client config, uses HTTP link

import { ApolloClient, InMemoryCache, createHttpLink, from } from '@apollo/client/core'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'

// @ctx: structured logging for Apollo operations (observability constraint)
const log = {
  info: (_msg: string, _ctx: Record<string, unknown>) => {},
  error: (msg: string, ctx: Record<string, unknown>) => {
    console.error(JSON.stringify({ level: 'error', msg, ...ctx, ts: new Date().toISOString() }))
  }
}

const httpLink = createHttpLink({
  uri: import.meta.env.VITE_GRAPHQL_ENDPOINT || '/graphql'
})

// @hlv:sec [AUTH_BOUNDARY] — JWT token attached to every GraphQL request
const authLink = setContext((_, { headers }) => {
  const token = localStorage.getItem('vedo-jwt-token')
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : ''
    }
  }
})

// @hlv log_all_errors — every Apollo error logged with full context
const errorLink = onError(({ graphQLErrors, networkError, operation }) => {
  if (graphQLErrors) {
    for (const err of graphQLErrors) {
      log.error('apollo.graphql_error', {
        message: err.message,
        locations: err.locations,
        path: err.path,
        operation: operation.operationName
      })
    }
  }
  if (networkError) {
    log.error('apollo.network_error', {
      message: networkError.message,
      operation: operation.operationName
    })
  }
})

export const apolloClient = new ApolloClient({
  link: from([errorLink, authLink, httpLink]),
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          ontology: {
            merge(_existing, incoming) {
              return incoming
            }
          }
        }
      }
    }
  }),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
      errorPolicy: 'all'
    },
    query: {
      fetchPolicy: 'cache-first',
      errorPolicy: 'all'
    }
  }
})

log.info('apollo.client.initialized', {
  endpoint: import.meta.env.VITE_GRAPHQL_ENDPOINT || '/graphql'
})
