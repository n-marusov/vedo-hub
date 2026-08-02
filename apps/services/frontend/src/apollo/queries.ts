// Apollo GraphQL queries for ontology graph navigation
//
// GraphQL in VEDO Core is restricted to ontology graph navigation only.
// All non-graph reads (versioning, org, comments, metrics, dashboard,
// deployments, merge requests) and all mutations migrated to REST (axios).
// See: ADR-DES.API.graphql-sparql-split-strategy.md

import { gql } from "@apollo/client/core";

// ── Fragments ───────────────────────────────────────────────────────────────────────

export const CLASS_SUMMARY_FRAGMENT = gql`
  fragment ClassSummaryFields on ClassSummary {
    id
    label
    comment
    parents
  }
`;

export const CLASS_FRAGMENT = gql`
  fragment ClassFields on Class {
    id
    label
    comment
    parents
    children
  }
`;

export const PROPERTY_FRAGMENT = gql`
  fragment PropertyFields on Property {
    id
    label
    comment
    propertyType
    domains
    ranges
    xsdType
    characteristics {
      functional
      inverseFunctional
      transitive
      symmetric
    }
    annotations {
      propertyIri
      value
    }
  }
`;

export const INDIVIDUAL_FRAGMENT = gql`
  fragment IndividualFields on Individual {
    id
    label
    comment
    classId
    classLabel
    literalValues {
      propertyId
      propertyLabel
      value
      xsdType
      valueId
    }
    referenceValues {
      propertyId
      propertyLabel
      targetId
      targetLabel
      edgeId
    }
  }
`;

// ── Class Queries ────────────────────────────────────────────────────────────────────

export const GET_CLASS_QUERY = gql`
  query GetClass($ontologyId: ID!, $classId: ID!) {
    class(ontologyId: $ontologyId, classId: $classId) {
      ...ClassFields
    }
  }
  ${CLASS_FRAGMENT}
`;

export const LIST_CLASSES_QUERY = gql`
  query ListClasses($ontologyId: ID!, $q: String, $page: Int, $perPage: Int) {
    classes(ontologyId: $ontologyId, q: $q, page: $page, perPage: $perPage) {
      items {
        ...ClassSummaryFields
      }
      total
      page
      perPage
    }
  }
  ${CLASS_SUMMARY_FRAGMENT}
`;

export const CLASS_TREE_QUERY = gql`
  query ClassTree($ontologyId: ID!) {
    classTree(ontologyId: $ontologyId) {
      id
      label
      comment
      children {
        id
        label
        comment
        children {
          id
          label
          comment
        }
      }
    }
  }
`;

export const CLASS_ANCESTORS_QUERY = gql`
  query ClassAncestors($ontologyId: ID!, $classId: ID!) {
    classAncestors(ontologyId: $ontologyId, classId: $classId) {
      id
      label
    }
  }
`;

export const CLASS_DESCENDANTS_QUERY = gql`
  query ClassDescendants($ontologyId: ID!, $classId: ID!, $maxDepth: Int) {
    classDescendants(
      ontologyId: $ontologyId
      classId: $classId
      maxDepth: $maxDepth
    ) {
      id
      label
      children {
        id
        label
      }
    }
  }
`;

export const GRAPH_NEIGHBORHOOD_QUERY = gql`
  query GraphNeighborhood(
    $ontologyId: ID!
    $classId: ID!
    $depth: Int
  ) {
    graphNeighborhood(
      ontologyId: $ontologyId
      classId: $classId
      depth: $depth
    ) {
      nodes {
        id
        label
      }
      edges {
        sourceId
        targetId
        propertyId
        propertyLabel
      }
    }
  }
`;

export const AUTOCOMPLETE_CLASSES_QUERY = gql`
  query AutocompleteClasses($ontologyId: ID!, $q: String!, $limit: Int) {
    autocompleteClasses(ontologyId: $ontologyId, q: $q, limit: $limit) {
      ...ClassSummaryFields
    }
  }
  ${CLASS_SUMMARY_FRAGMENT}
`;

// ── Property Queries ─────────────────────────────────────────────────────────────────

export const GET_PROPERTY_QUERY = gql`
  query GetProperty($ontologyId: ID!, $propertyId: ID!) {
    property(ontologyId: $ontologyId, propertyId: $propertyId) {
      ...PropertyFields
    }
  }
  ${PROPERTY_FRAGMENT}
`;

export const LIST_PROPERTIES_QUERY = gql`
  query ListProperties(
    $ontologyId: ID!
    $q: String
    $propertyType: PropertyType
    $page: Int
    $perPage: Int
  ) {
    properties(
      ontologyId: $ontologyId
      q: $q
      propertyType: $propertyType
      page: $page
      perPage: $perPage
    ) {
      items {
        id
        label
        propertyType
        xsdType
        domains
      }
      total
      page
      perPage
    }
  }
`;

// ── Individual Queries ───────────────────────────────────────────────────────────────

export const GET_INDIVIDUAL_QUERY = gql`
  query GetIndividual($ontologyId: ID!, $individualId: ID!) {
    individual(ontologyId: $ontologyId, individualId: $individualId) {
      ...IndividualFields
    }
  }
  ${INDIVIDUAL_FRAGMENT}
`;

export const LIST_INDIVIDUALS_QUERY = gql`
  query ListIndividuals(
    $ontologyId: ID!
    $classId: ID!
    $q: String
    $page: Int
    $perPage: Int
  ) {
    individuals(
      ontologyId: $ontologyId
      classId: $classId
      q: $q
      page: $page
      perPage: $perPage
    ) {
      items {
        ...IndividualFields
      }
      total
      page
      perPage
    }
  }
  ${INDIVIDUAL_FRAGMENT}
`;
