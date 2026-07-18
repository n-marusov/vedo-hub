// Apollo GraphQL queries and mutations for ontology workspace
//
// These queries match the async-graphql schema exposed by ontology-service
// at /api/v1/graphql via the API Gateway proxy.

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

// ── Queries ─────────────────────────────────────────────────────────────────────────

/// Ontology metadata — branch, commit, dirty state
export const ONTOLOGY_QUERY = gql`
  query Ontology($id: ID!) {
    ontology(id: $id) {
      id
      name
      branch
      commit
      dirty
    }
  }
`;

/// Version context — used by the toolbar to show current branch/commit
export const VERSION_CONTEXT_QUERY = gql`
  query VersionContext($id: ID!) {
    ontology(id: $id) {
      branch
      commit
      dirty
    }
  }
`;

/// Single class with full details
export const GET_CLASS_QUERY = gql`
  query GetClass($ontologyId: ID!, $classId: ID!) {
    class(ontologyId: $ontologyId, classId: $classId) {
      ...ClassFields
    }
  }
  ${CLASS_FRAGMENT}
`;

/// Paginated class list with search
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

/// Class hierarchy tree (root classes)
export const CLASS_TREE_QUERY = gql`
  query ClassTree($ontologyId: ID!) {
    classTree(ontologyId: $ontologyId) {
      id
      label
      children {
        id
        label
        children {
          id
          label
        }
      }
    }
  }
`;

/// Class ancestors (breadcrumb path)
export const CLASS_ANCESTORS_QUERY = gql`
  query ClassAncestors($ontologyId: ID!, $classId: ID!) {
    classAncestors(ontologyId: $ontologyId, classId: $classId) {
      id
      label
    }
  }
`;

/// Class descendants tree
export const CLASS_DESCENDANTS_QUERY = gql`
  query ClassDescendants($ontologyId: ID!, $classId: ID!, $maxDepth: Int) {
    classDescendants(ontologyId: $ontologyId, classId: $classId, maxDepth: $maxDepth) {
      id
      label
      children {
        id
        label
      }
    }
  }
`;

/// Graph neighborhood for a class
export const GRAPH_NEIGHBORHOOD_QUERY = gql`
  query GraphNeighborhood($ontologyId: ID!, $classId: ID!, $depth: Int) {
    graphNeighborhood(ontologyId: $ontologyId, classId: $classId, depth: $depth) {
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

/// Autocomplete class search
export const AUTOCOMPLETE_CLASSES_QUERY = gql`
  query AutocompleteClasses($ontologyId: ID!, $q: String!, $limit: Int) {
    autocompleteClasses(ontologyId: $ontologyId, q: $q, limit: $limit) {
      ...ClassSummaryFields
    }
  }
  ${CLASS_SUMMARY_FRAGMENT}
`;

/// Single property with full details
export const GET_PROPERTY_QUERY = gql`
  query GetProperty($ontologyId: ID!, $propertyId: ID!) {
    property(ontologyId: $ontologyId, propertyId: $propertyId) {
      ...PropertyFields
    }
  }
  ${PROPERTY_FRAGMENT}
`;

/// Paginated property list with type filter
export const LIST_PROPERTIES_QUERY = gql`
  query ListProperties($ontologyId: ID!, $q: String, $propertyType: PropertyType, $page: Int, $perPage: Int) {
    properties(ontologyId: $ontologyId, q: $q, propertyType: $propertyType, page: $page, perPage: $perPage) {
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

/// Single individual with full detail (property values)
export const GET_INDIVIDUAL_QUERY = gql`
  query GetIndividual($ontologyId: ID!, $individualId: ID!) {
    individual(ontologyId: $ontologyId, individualId: $individualId) {
      ...IndividualFields
    }
  }
  ${INDIVIDUAL_FRAGMENT}
`;

/// Paginated individual list filtered by class
export const LIST_INDIVIDUALS_QUERY = gql`
  query ListIndividuals($ontologyId: ID!, $classId: ID!, $q: String, $page: Int, $perPage: Int) {
    individuals(ontologyId: $ontologyId, classId: $classId, q: $q, page: $page, perPage: $perPage) {
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
`;

// ── Mutations ───────────────────────────────────────────────────────────────────────

/// Update draft state — marks workspace as dirty
export const UPDATE_DRAFT_MUTATION = gql`
  mutation UpdateDraft($ontologyId: ID!, $changes: DraftInput!) {
    updateDraft(ontologyId: $ontologyId, changes: $changes) {
      success
      timestamp
    }
  }
`;

// ── Versioning Queries (Commit / Branch) ──────────────────────────────────────────────

/// Commit summary fields reused by history and diff views.
export const COMMIT_SUMMARY_FRAGMENT = gql`
  fragment CommitSummaryFields on Commit {
    id
    branchId
    parentCommitId
    message
    authorId
    authorName
    totalChanges
    createdAt
  }
`;

/// Paginated commit history for a branch.
/// Backed by the ontology-service GraphQL resolver which proxies the
/// versioning-service `/api/v1/versioning/commits` endpoint.
export const GET_COMMIT_HISTORY_QUERY = gql`
  query GetCommitHistory(
    $ontologyId: ID!
    $branchId: ID
    $page: Int
    $perPage: Int
  ) {
    commits(
      ontologyId: $ontologyId
      branchId: $branchId
      page: $page
      perPage: $perPage
    ) {
      items {
        ...CommitSummaryFields
      }
      total
      page
      perPage
    }
  }
  ${COMMIT_SUMMARY_FRAGMENT}
`;

/// Branch summary fields.
export const BRANCH_FRAGMENT = gql`
  fragment BranchFields on Branch {
    id
    name
    ontologyId
    headCommitId
    createdAt
    isProtected
    lastCommitMessage
    lastCommitAuthor
    aheadCount
    behindCount
  }
`;

/// List branches for an ontology with optional reference branch for ahead/behind.
export const GET_BRANCHES_QUERY = gql`
  query GetBranches(
    $ontologyId: ID!
    $referenceBranchId: ID
  ) {
    branches(
      ontologyId: $ontologyId
      referenceBranchId: $referenceBranchId
    ) {
      items {
        ...BranchFields
      }
      total
    }
  }
  ${BRANCH_FRAGMENT}
`;

/// Single branch by ID (within an ontology).
export const GET_BRANCH_QUERY = gql`
  query GetBranch($ontologyId: ID!, $branchId: ID!) {
    branch(ontologyId: $ontologyId, branchId: $branchId) {
      ...BranchFields
    }
  }
  ${BRANCH_FRAGMENT}
`;

/// Navigation state query
export const NAVIGATION_STATE_QUERY = gql`
  query NavigationState {
    userPreferences {
      sidebarCollapsed
      activeRoute
      theme
    }
  }
`;

// ── Comment Query ────────────────────────────────────────────────────────────

/// Comment fields fragment.
export const COMMENT_FRAGMENT = gql`
  fragment CommentFields on Comment {
    id
    author
    authorName
    text
    entityId
    entityType
    parentCommentId
    createdAt
    updatedAt
  }
`;

/// Fetch comments for an entity (class/property/individual).
export const GET_ENTITY_COMMENTS_QUERY = gql`
  query GetEntityComments($ontologyId: ID!, $entityId: ID!) {
    comments(ontologyId: $ontologyId, entityId: $entityId) {
      ...CommentFields
    }
  }
  ${COMMENT_FRAGMENT}
`;

/// Fetch project-wide comment feed.
export const GET_COMMENT_FEED_QUERY = gql`
  query GetCommentFeed($ontologyId: ID!, $page: Int, $perPage: Int) {
    commentFeed(ontologyId: $ontologyId, page: $page, perPage: $perPage) {
      items {
        ...CommentFields
      }
      total
      page
      perPage
    }
  }
  ${COMMENT_FRAGMENT}
`;

// ── M2.5 Queries & Mutations ──────────────────────────────────────────────────────
// @m2.5 — Added for GUI wiring implementation

/// Execute a SPARQL query against the ontology.
export const SPARQL_EXECUTE_QUERY = gql`
  query SparqlExecute($ontologyId: ID!, $query: String!, $limit: Int, $offset: Int) {
    sparqlQuery(ontologyId: $ontologyId, query: $query, limit: $limit, offset: $offset)
      columns
      rows
      total
      executionTimeMs
  }
`;

/// List projects with search, sort, pagination.
export const LIST_PROJECTS_QUERY = gql`
  query ListProjects($q: String, $sortBy: String, $sortDir: SortDirection, $page: Int, $perPage: Int) {
    projects(q: $q, sortBy: $sortBy, sortDir: $sortDir, page: $page, perPage: $perPage)
      items
        id
        name
        description
        visibility
        ontologyCount
        memberCount
        updatedAt
      total
      page
      perPage
  }
`;

/// List groups with hierarchy and search.
export const LIST_GROUPS_QUERY = gql`
  query ListGroups($q: String) {
    groups(q: $q)
      id
      name
      description
      parentGroupId
      childGroups
        id
        name
      memberCount
      projectCount
  }
`;

/// List members of an ontology with roles.
export const LIST_MEMBERS_QUERY = gql`
  query ListMembers($ontologyId: ID!) {
    members(ontologyId: $ontologyId)
      id
      userId
      username
      avatarUrl
      role
      addedAt
  }
`;

/// Update a member's role.
export const UPDATE_MEMBER_ROLE_MUTATION = gql`
  mutation UpdateMemberRole($ontologyId: ID!, $userId: ID!, $role: String!) {
    updateMemberRole(ontologyId: $ontologyId, userId: $userId, role: $role)
      success
      member
        id
        userId
        role
  }
`;

/// Remove a member from an ontology.
export const REMOVE_MEMBER_MUTATION = gql`
  mutation RemoveMember($ontologyId: ID!, $userId: ID!) {
    removeMember(ontologyId: $ontologyId, userId: $userId)
      success
  }
`;

/// Get tags for a versioning context.
export const GET_TAGS_QUERY = gql`
  query GetTags($ontologyId: ID!) {
    tags(ontologyId: $ontologyId)
      id
      name
      commitId
      message
      authorName
      createdAt
  }
`;

/// Compare two revisions and return diff data.
export const COMPARE_REVISIONS_QUERY = gql`
  query CompareRevisions($ontologyId: ID!, $fromRevision: ID!, $toRevision: ID!) {
    compareRevisions(ontologyId: $ontologyId, fromRevision: $fromRevision, toRevision: $toRevision)
      additions
      deletions
      changes
        entityId
        entityType
        entityLabel
        changeType
        field
        oldValue
        newValue
  }
`;

/// Dashboard aggregate query — widgets, recent ontologies, activity feed.
export const DASHBOARD_QUERY = gql`
  query DashboardAggregate {
    dashboard
      widgets
        title
        count
        icon
        route
      recentOntologies
        id
        name
        description
        visibility
        updatedAt
      activityFeed
        id
        text
        author
        timestamp
        type
      attentionItems
        id
        text
        severity
        count
  }
`;

/// Ontology metrics — KPI counters and trends.
export const ONTOLOGY_METRICS_QUERY = gql`
  query OntologyMetrics($ontologyId: ID!) {
    ontologyMetrics(ontologyId: $ontologyId)
      counters
        classCount
        propertyCount
        individualCount
        axiomCount
        commentCount
        mergeRequestCount
      trends
        date
        classCount
        propertyCount
        individualCount
  }
`;

/// Run SHACL validation (currently returns OK stub).
export const RUN_VALIDATION_MUTATION = gql`
  mutation RunValidation($ontologyId: ID!) {
    runValidation(ontologyId: $ontologyId)
      status
      violations
        path
        message
        severity
        node
      validatedAt
  }
`;

/// List deployments.
export const LIST_DEPLOYMENTS_QUERY = gql`
  query ListDeployments($includeStopped: Boolean) {
    deployments(includeStopped: $includeStopped)
      id
      url
      status
      version
      ontologyId
      ontologyName
      deployedAt
      deployedBy
  }
`;

/// List merge requests with sections and tabs.
export const LIST_MERGE_REQUESTS_QUERY = gql`
  query ListMergeRequests($status: String) {
    mergeRequests(status: $status)
      id
      title
      description
      sourceBranch
      targetBranch
      authorName
      status
      mergeStatus
      createdAt
      commentCount
  }
`;
