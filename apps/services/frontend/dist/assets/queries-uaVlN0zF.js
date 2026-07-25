import{g as e}from"./index-BxgMCe-w.js";const o=e`
  fragment ClassSummaryFields on ClassSummary {
    id
    label
    comment
    parents
  }
`,l=e`
  fragment ClassFields on Class {
    id
    label
    comment
    parents
    children
  }
`,t=e`
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
`,s=e`
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
`;e`
  query GetClass($ontologyId: ID!, $classId: ID!) {
    class(ontologyId: $ontologyId, classId: $classId) {
      ...ClassFields
    }
  }
  ${l}
`;e`
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
  ${o}
`;const d=e`
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
`;e`
  query ClassAncestors($ontologyId: ID!, $classId: ID!) {
    classAncestors(ontologyId: $ontologyId, classId: $classId) {
      id
      label
    }
  }
`;e`
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
`;const r=e`
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
`;e`
  query AutocompleteClasses($ontologyId: ID!, $q: String!, $limit: Int) {
    autocompleteClasses(ontologyId: $ontologyId, q: $q, limit: $limit) {
      ...ClassSummaryFields
    }
  }
  ${o}
`;e`
  query GetProperty($ontologyId: ID!, $propertyId: ID!) {
    property(ontologyId: $ontologyId, propertyId: $propertyId) {
      ...PropertyFields
    }
  }
  ${t}
`;e`
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
`;e`
  query GetIndividual($ontologyId: ID!, $individualId: ID!) {
    individual(ontologyId: $ontologyId, individualId: $individualId) {
      ...IndividualFields
    }
  }
  ${s}
`;const n=e`
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
        id
        label
        comment
        classId
        classLabel
      }
      total
      page
      perPage
    }
  }
`;export{d as C,r as G,n as L};
