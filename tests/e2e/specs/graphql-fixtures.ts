// E2E test fixtures — mock auth, deterministic GUI GraphQL data
//
// IN REAL-BACKEND MODE:
// Uses a pre-signed RS256 JWT (OWNER_JWT) that matches the test public key
// configured in deploy/docker-compose.test.yml (JWT_DEV_PUBLIC_KEY_PEM).
// The matching private key is in test-jwt-key.pem.
//
// GUI specs verify page wiring and UX states for dashboard-level GraphQL
// fields that are intentionally not part of the current real ontology-service
// schema. This fixture supplies deterministic GUI data for those operations and
// falls through to the real backend for every unknown operation.
import { test as base, expect, type Route } from '@playwright/test';

// Pre-signed RS256 JWT with Owner role — matches api-gateway-test's public key
import { OWNER_JWT } from './jwt-tokens';

const MOCK_DASHBOARD = {
  widgets: [
    { title: 'Merge Requests', count: 3, icon: 'git-merge', route: '/merge-requests' },
    { title: 'Reviews', count: 2, icon: 'eye', route: '/reviews' },
    { title: 'Work Items', count: 7, icon: 'list-todo', route: '/work-items' },
  ],
  attentionItems: [
    { id: 'att-001', text: 'Merge request MR-001 needs review', severity: 'warning', count: 1 },
    { id: 'att-002', text: 'Version conflict in ontology', severity: 'error', count: 2 },
  ],
  activityFeed: [
    { id: 'act-001', text: 'owner_seed created University Ontology', author: 'owner_seed', timestamp: '2026-03-15T14:30:00Z', type: 'ontology' },
    { id: 'act-002', text: 'editor_seed opened merge request MR-001', author: 'editor_seed', timestamp: '2026-03-15T12:00:00Z', type: 'merge_request' },
  ],
  recentOntologies: [
    { id: 'ont-123', name: 'University Ontology', description: 'Academic ontology project', visibility: 'private', updatedAt: '2026-03-15T14:30:00Z' },
    { id: 'ont-456', name: 'Healthcare Terms', description: 'Medical terminology ontology', visibility: 'internal', updatedAt: '2026-03-14T12:00:00Z' },
  ],
};

const MOCK_PROJECTS = {
  items: [
    { id: 'ont-123', name: 'Test University Ontology', description: 'Academic ontology project', visibility: 'private', ontologyCount: 3, memberCount: 5, updatedAt: '2026-03-15T14:30:00Z' },
    { id: 'ont-456', name: 'Healthcare Terms', description: 'Medical terminology project', visibility: 'internal', ontologyCount: 2, memberCount: 8, updatedAt: '2026-03-14T12:00:00Z' },
    { id: 'ont-789', name: 'Financial Taxonomy', description: 'Finance project', visibility: 'public', ontologyCount: 1, memberCount: 3, updatedAt: '2026-03-13T10:00:00Z' },
  ],
  total: 3,
  page: 1,
  perPage: 50,
};

const MOCK_GROUPS = {
  items: [
    { id: 'grp-001', name: 'Engineering', description: 'Engineering team', visibility: 'private', parentGroupId: null, childGroups: [{ id: 'grp-002', name: 'Data Science', description: 'Data science team', visibility: 'private', parentGroupId: 'grp-001', childGroups: [], memberCount: 8, projectCount: 3 }], memberCount: 12, projectCount: 5 },
    { id: 'grp-003', name: 'Research', description: 'Research division', visibility: 'public', parentGroupId: null, childGroups: [], memberCount: 5, projectCount: 2 },
  ],
};

const MOCK_MEMBERS = [
  { id: 'user-456', userId: 'user-456', username: 'owner_seed', avatarUrl: '', role: 'Owner', addedAt: '2026-01-15T10:00:00Z' },
  { id: 'user-789', userId: 'user-789', username: 'editor_seed', avatarUrl: '', role: 'Editor', addedAt: '2026-02-20T14:30:00Z' },
  { id: 'user-012', userId: 'user-012', username: 'viewer_seed', avatarUrl: '', role: 'Viewer', addedAt: '2026-03-10T09:15:00Z' },
];

const MOCK_METRICS = {
  counters: { classCount: 156, propertyCount: 89, individualCount: 1204, axiomCount: 18450, commentCount: 42, mergeRequestCount: 3 },
  trends: [
    { date: '2026-03-01', classCount: 140, propertyCount: 80, individualCount: 1100 },
    { date: '2026-03-08', classCount: 148, propertyCount: 85, individualCount: 1150 },
    { date: '2026-03-15', classCount: 156, propertyCount: 89, individualCount: 1204 },
  ],
};

const MOCK_DEPLOYMENTS = [
  { id: 'dep-001', url: 'https://example.com/onto/1', status: 'active', version: '1.0', ontologyId: 'ont-123', ontologyName: 'University Ontology', deployedAt: '2026-02-01T10:00:00Z', deployedBy: 'owner_seed' },
  { id: 'dep-002', url: 'https://example.com/onto/2', status: 'stopped', version: '2.0', ontologyId: 'ont-456', ontologyName: 'Healthcare Terms', deployedAt: '2026-01-15T10:00:00Z', deployedBy: 'editor_seed' },
];

const MOCK_MERGE_REQUESTS = [
  { id: 'mr-001', title: 'Add Person class hierarchy', description: 'Adding Person class with subclasses', sourceBranch: 'feature/person-hierarchy', targetBranch: 'main', authorName: 'editor_seed', status: 'open', mergeStatus: 'unchecked', createdAt: '2026-03-14T10:00:00Z', commentCount: 2 },
  { id: 'mr-002', title: 'Fix property constraints', description: 'Fix property range constraints', sourceBranch: 'fix/property-constraints', targetBranch: 'main', authorName: 'owner_seed', status: 'merged', mergeStatus: 'merged', createdAt: '2026-03-10T10:00:00Z', commentCount: 0 },
];

const MOCK_COMMITS = {
  items: [
    { __typename: 'Commit', id: 'commit-001', branchId: 'branch-main', parentCommitId: 'commit-000', message: 'Initial ontology setup', authorId: 'user-456', authorName: 'owner_seed', totalChanges: 15, createdAt: '2026-03-01T10:00:00Z' },
    { __typename: 'Commit', id: 'commit-002', branchId: 'branch-main', parentCommitId: 'commit-001', message: 'Add Person class', authorId: 'user-789', authorName: 'editor_seed', totalChanges: 3, createdAt: '2026-03-05T14:00:00Z' },
    { __typename: 'Commit', id: 'commit-003', branchId: 'branch-main', parentCommitId: 'commit-002', message: 'Add properties', authorId: 'user-789', authorName: 'editor_seed', totalChanges: 8, createdAt: '2026-03-10T09:00:00Z' },
  ],
  total: 3,
  page: 1,
  perPage: 50,
};

const MOCK_BRANCHES = {
  items: [
    { __typename: 'Branch', id: 'branch-main', name: 'main', ontologyId: 'ont-123', headCommitId: 'commit-003', createdAt: '2026-01-15T10:00:00Z', isProtected: true, lastCommitMessage: 'Add properties', lastCommitAuthor: 'editor_seed', aheadCount: 0, behindCount: 0 },
    { __typename: 'Branch', id: 'branch-dev', name: 'develop', ontologyId: 'ont-123', headCommitId: 'commit-002', createdAt: '2026-02-01T10:00:00Z', isProtected: false, lastCommitMessage: 'Add Person class', lastCommitAuthor: 'editor_seed', aheadCount: 1, behindCount: 0 },
  ],
  total: 2,
};

const MOCK_VERSIONING_COMMITS = {
  items: [
    { id: 'commit-001', branchId: 'branch-main', parentCommitId: null, message: 'Initial ontology setup', authorId: 'user-456', authorName: 'owner_seed', totalChanges: 15, createdAt: '2026-03-01T10:00:00Z' },
    { id: 'commit-002', branchId: 'branch-main', parentCommitId: 'commit-001', message: 'Add Person class', authorId: 'user-789', authorName: 'editor_seed', totalChanges: 3, createdAt: '2026-03-05T14:00:00Z' },
  ],
  total: 2,
  page: 0,
  perPage: 50,
};

const MOCK_VERSIONING_BRANCHES = {
  items: [
    { id: 'branch-main', name: 'main', ontologyId: 'ont-123', headCommitId: 'commit-002', createdAt: '2026-01-15T10:00:00Z', isProtected: true, lastCommitMessage: 'Add Person class', lastCommitAuthor: 'editor_seed', aheadCount: 0, behindCount: 0 },
    { id: 'branch-dev', name: 'develop', ontologyId: 'ont-123', headCommitId: 'commit-001', createdAt: '2026-02-01T10:00:00Z', isProtected: false, lastCommitMessage: 'Initial ontology setup', lastCommitAuthor: 'owner_seed', aheadCount: 1, behindCount: 0 },
  ],
  total: 2,
};

const MOCK_VERSIONING_TAGS = [
  { id: 'tag-001', name: 'v1.0.0', commitId: 'commit-002', message: 'Initial release', authorName: 'owner_seed', createdAt: '2026-03-15T10:00:00Z' },
  { id: 'tag-002', name: 'v0.9.0', commitId: 'commit-001', message: 'Beta', authorName: 'editor_seed', createdAt: '2026-03-10T10:00:00Z' },
];

const MOCK_TAGS = [
  { id: 'tag-001', name: 'v1.0.0', commitId: 'commit-003', message: 'Initial release', authorName: 'owner_seed', createdAt: '2026-03-15T10:00:00Z' },
  { id: 'tag-002', name: 'v0.9.0', commitId: 'commit-002', message: 'Beta', authorName: 'editor_seed', createdAt: '2026-03-10T10:00:00Z' },
];

const MOCK_GRAPH_NEIGHBORHOOD = {
  nodes: [
    { id: 'commit-001', label: 'Initial ontology setup' },
    { id: 'commit-002', label: 'Add Person class' },
    { id: 'commit-003', label: 'Add properties' },
  ],
  edges: [
    { sourceId: 'commit-001', targetId: 'commit-002', propertyId: 'parent', propertyLabel: 'parent' },
    { sourceId: 'commit-002', targetId: 'commit-003', propertyId: 'parent', propertyLabel: 'parent' },
  ],
};

// Mock ontology workspace class tree data — used by workspace's CLASS_TREE_QUERY
// The graph nodes are computed from the class tree, not from a separate graph query.
const MOCK_CLASS_TREE = [
  {
    id: 'Person',
    label: 'Person',
    comment: 'A person',
    children: [
      { id: 'Student', label: 'Student', comment: 'A student', children: [] },
      { id: 'Professor', label: 'Professor', comment: 'A professor', children: [] },
    ],
  },
  { id: 'Organization', label: 'Organization', comment: 'An organization', children: [] },
];

// Mock individuals data — used for graph edge generation (individual → class links)
const MOCK_INDIVIDUALS = {
  items: [
    { id: 'John', label: 'John', comment: '', classId: 'Professor', classLabel: 'Professor' },
    { id: 'Alice', label: 'Alice', comment: '', classId: 'Student', classLabel: 'Student' },
  ],
  total: 2,
  page: 0,
  perPage: 50,
};

const MOCK_COMPARE_REVISIONS = {
  additions: 5,
  deletions: 3,
  changes: [
    { entityId: 'cls-001', entityType: 'class', entityLabel: 'Person', changeType: 'added', field: 'label', oldValue: null, newValue: 'Person' },
    { entityId: 'cls-002', entityType: 'class', entityLabel: 'Student', changeType: 'added', field: 'label', oldValue: null, newValue: 'Student' },
  ],
};

const MOCK_SPARQL_RESULTS = {
  columns: ['s', 'p', 'o'],
  rows: [
    ['ex:Person', 'rdf:type', 'owl:Class'],
    ['ex:Student', 'rdfs:subClassOf', 'ex:Person'],
  ],
  total: 2,
  executionTimeMs: 15,
};

const MOCK_VALIDATION = { status: 'ok', violations: [], validatedAt: '2026-03-15T14:30:00Z' };

// Create a mock session for E2E testing without Keycloak
function createMockSession() {
  return {
    accessToken: OWNER_JWT,
    refreshToken: OWNER_JWT,
    userId: 'user-123',
    tenantId: 'default',
    roles: ['owner'],
    expiresAt: Date.now() + 86400000,
  };
}

async function fulfillGraphql(route: Route, data: unknown, delayMs = 0) {
  if (delayMs > 0) {
    await new Promise((resolve) => setTimeout(resolve, delayMs));
  }

  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ data }),
  });
}

async function handleGuiGraphql(route: Route) {
  const request = route.request();
  let body: { operationName?: string; query?: string; variables?: Record<string, unknown> } = {};

  try {
    body = request.postDataJSON();
  } catch {
    return route.fallback();
  }

  const query = body.query || '';
  const operationName = body.operationName || '';
  const signature = `${operationName}\n${query}`;

  if (signature.includes('DashboardAggregate') || signature.includes('dashboard')) {
    return fulfillGraphql(route, { dashboard: MOCK_DASHBOARD });
  }

  if (signature.includes('ListProjects') || signature.includes('projects')) {
    return fulfillGraphql(route, { projects: MOCK_PROJECTS });
  }

  if (signature.includes('ListGroups') || signature.includes('groups')) {
    return fulfillGraphql(route, { groups: MOCK_GROUPS });
  }

  if (signature.includes('ListMembers') || signature.includes('members')) {
    return fulfillGraphql(route, { members: MOCK_MEMBERS });
  }

  if (signature.includes('OntologyMetrics') || signature.includes('ontologyMetrics')) {
    return fulfillGraphql(route, { ontologyMetrics: MOCK_METRICS }, 350);
  }

  if (signature.includes('ListDeployments') || signature.includes('deployments')) {
    return fulfillGraphql(route, { deployments: MOCK_DEPLOYMENTS });
  }

  if (signature.includes('ListMergeRequests') || signature.includes('mergeRequests')) {
    return fulfillGraphql(route, { mergeRequests: MOCK_MERGE_REQUESTS });
  }

  if (signature.includes('GetCommitHistory') || signature.includes('commits')) {
    return fulfillGraphql(route, { commits: MOCK_COMMITS }, 200);
  }

  if (signature.includes('GetBranches') || signature.includes('branches')) {
    return fulfillGraphql(route, { branches: MOCK_BRANCHES }, 200);
  }

  if (signature.includes('GetTags') || signature.includes('tags')) {
    return fulfillGraphql(route, { tags: MOCK_TAGS }, 200);
  }

  if (signature.includes('CompareRevisions') || signature.includes('compareRevisions')) {
    return fulfillGraphql(route, { compareRevisions: MOCK_COMPARE_REVISIONS }, 200);
  }

  if (signature.includes('GraphNeighborhood') || signature.includes('graphNeighborhood')) {
    return fulfillGraphql(route, { graphNeighborhood: MOCK_GRAPH_NEIGHBORHOOD }, 200);
  }

  	if (signature.includes('RunValidation') || signature.includes('runValidation')) {
    return fulfillGraphql(route, { runValidation: MOCK_VALIDATION }, 300);
  }

  if (signature.includes('UpdateMemberRole') || signature.includes('updateMemberRole')) {
    return fulfillGraphql(route, {
      updateMemberRole: {
        success: true,
        member: { id: body.variables?.userId || 'user-789', userId: body.variables?.userId || 'user-789', role: body.variables?.role || 'Editor' },
      },
    });
  }

  if (signature.includes('RemoveMember') || signature.includes('removeMember')) {
    return fulfillGraphql(route, { removeMember: { success: true } });
  }

  if (signature.includes('ClassTree') || signature.includes('classTree')) {
    return fulfillGraphql(route, { classTree: MOCK_CLASS_TREE }, 200);
  }

  if (signature.includes('ListIndividuals') || signature.includes('listIndividuals')) {
    return fulfillGraphql(route, { individuals: MOCK_INDIVIDUALS }, 200);
  }

  return route.fallback();
}

async function handleSparqlRest(route: Route) {
  const request = route.request();
  let body: { query?: string } = {};
  try {
    body = request.postDataJSON();
  } catch {
    return route.fallback();
  }

  const query = body.query || '';
  if (/invalid/i.test(query)) {
    return route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ error: { code: 'SPARQL-SYNTAX-ERROR', message: 'SPARQL syntax error near INVALID' } }),
    });
  }

  await new Promise((resolve) => setTimeout(resolve, 300));
  return route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify(MOCK_SPARQL_RESULTS),
  });
}

export const test = base.extend({
  page: async ({ page }, use) => {
    const session = createMockSession();
    const ownerJwt = OWNER_JWT;

    await page.addInitScript(({ session, ownerJwt }) => {
      // Set auth session for router guard (sessionStorage key: vedo_session)
      sessionStorage.setItem('vedo_session', JSON.stringify(session));

      // Set JWT for Apollo Client and REST API (localStorage key: vedo-jwt-token)
      // This token is sent as Authorization: Bearer <token> on every API request.
      localStorage.setItem('vedo-jwt-token', ownerJwt);
    }, { session, ownerJwt });

    await page.route('**/api/v1/graphql', handleGuiGraphql);
    await page.route('**/graphql', handleGuiGraphql);
    await page.route('**/api/v1/sparql', handleSparqlRest);
    await page.route('**/api/v1/groups*', handleRestGroups);
    await page.route('**/api/v1/versioning/*', handleRestVersioning);
    await page.route('**/api/v1/ontologies/*/validate', handleValidationRest);

    await use(page);
  },
});

export { expect } from '@playwright/test';

const MOCK_GROUP_RESPONSE = {
  data: [
    {
      id: 'grp-001', slug: 'engineering', name: 'Engineering', description: 'Engineering team',
      visibility: 'private', parent_id: null,
      member_count: 12, project_count: 5,
    },
    {
      id: 'grp-002', slug: 'data-science', name: 'Data Science', description: 'Data science team',
      visibility: 'private', parent_id: 'grp-001',
      member_count: 8, project_count: 3,
    },
    {
      id: 'grp-003', slug: 'research', name: 'Research', description: 'Research division',
      visibility: 'public', parent_id: null,
      member_count: 5, project_count: 2,
    },
  ],
};

let groupStore: { data: any[] } = JSON.parse(JSON.stringify(MOCK_GROUP_RESPONSE));

async function handleRestGroups(route: Route) {
  const request = route.request();

  if (request.method() === 'GET') {
    const url = new URL(request.url());
    const search = url.searchParams.get('search');
    let groups = groupStore.data;
    if (search) {
      groups = groups.filter((g: any) =>
        g.name.toLowerCase().includes(search.toLowerCase())
      );
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: groups }),
    });
  }

  if (request.method() === 'POST') {
    try {
      const body = request.postDataJSON();
      const newGroup = {
        id: 'grp-new-' + Date.now(),
        slug: body.slug || String(body.label || body.name || 'group').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, ''),
        name: body.label || body.name,
        description: body.description || '',
        visibility: body.visibility || 'private',
        parent_id: body.parent_id || null,
        member_count: 1,
        project_count: 0,
      };
      groupStore.data.push(newGroup);
      return route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ data: newGroup }),
      });
    } catch {
      return route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ error: { code: 'INVALID_REQUEST', message: 'Invalid request body' } }),
      });
    }
  }

  return route.fallback();
}

async function handleRestVersioning(route: Route) {
  const request = route.request();
  const url = new URL(request.url());
  const path = url.pathname;

  if (request.method() !== 'GET') return route.fallback();

  if (path.endsWith('/commits')) {
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(MOCK_VERSIONING_COMMITS),
    });
  }

  if (path.endsWith('/branches')) {
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(MOCK_VERSIONING_BRANCHES),
    });
  }

  if (path.endsWith('/tags')) {
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(MOCK_VERSIONING_TAGS),
    });
  }

  return route.fallback();
}

async function handleValidationRest(route: Route) {
  const request = route.request();
  if (request.method() !== 'POST') return route.fallback();

  await new Promise((resolve) => setTimeout(resolve, 700));
  return route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify(MOCK_VALIDATION),
  });
}
