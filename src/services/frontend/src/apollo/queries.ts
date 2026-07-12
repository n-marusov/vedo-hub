// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Apollo GraphQL queries and mutations for ontology workspace

import { gql } from '@apollo/client/core'

export const ONTOLOGY_QUERY = gql`
  query Ontology($ontologyId: ID!) {
    ontology(id: $ontologyId) {
      id
      name
      branch
      commit
      dirty
      classes {
        id
        label
        childrenCount
      }
      properties {
        id
        label
        type
      }
    }
  }
`

export const VERSION_CONTEXT_QUERY = gql`
  query VersionContext($ontologyId: ID!) {
    ontology(id: $ontologyId) {
      branch
      commit
      dirty
    }
  }
`

export const UPDATE_DRAFT_MUTATION = gql`
  mutation UpdateDraft($ontologyId: ID!, $changes: DraftInput!) {
    updateDraft(ontologyId: $ontologyId, changes: $changes) {
      success
      timestamp
    }
  }
`

export const NAVIGATION_STATE_QUERY = gql`
  query NavigationState {
    userPreferences {
      sidebarCollapsed
      activeRoute
      theme
    }
  }
`
