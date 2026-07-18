// @ctx: M2.5 E2E stub API server — returns mock JSON responses for all M2.5 endpoints
// Started alongside frontend Vite dev server for E2E testing without Docker

import http from 'node:http';

const PORT = 3001;

// Mock data constants
const MOCK_ONTOLOGIES = [
  { id: 'ont-001', name: 'University Ontology', visibility: 'private', createdAt: '2026-01-15T10:00:00Z' },
  { id: 'ont-002', name: 'Healthcare Terms', visibility: 'internal', createdAt: '2026-02-20T14:30:00Z' },
  { id: 'ont-003', name: 'Financial Taxonomy', visibility: 'public', createdAt: '2026-03-10T09:15:00Z' },
];

const MOCK_GROUPS = [
  { id: 'grp-001', name: 'Engineering', memberCount: 12, parentId: null },
  { id: 'grp-002', name: 'Data Science', memberCount: 8, parentId: 'grp-001' },
  { id: 'grp-003', name: 'Research', memberCount: 5, parentId: null },
];

const MOCK_MEMBERS = [
  { id: 'user-456', username: 'owner_seed', role: 'owner', email: 'owner@vedo.dev' },
  { id: 'user-789', username: 'editor_seed', role: 'editor', email: 'editor@vedo.dev' },
  { id: 'user-012', username: 'viewer_seed', role: 'viewer', email: 'viewer@vedo.dev' },
];

const MOCK_TAGS = [
  { id: 'tag-001', name: 'v1.0.0', revision: 'abc123', createdAt: '2026-01-20T10:00:00Z' },
  { id: 'tag-002', name: 'v2.0.0', revision: 'def456', createdAt: '2026-03-15T14:00:00Z' },
];

const MOCK_METRICS = {
  counters: { classes: 156, properties: 89, individuals: 1204, axioms: 18450 },
  trends: [
    { date: '2026-03-01', classCount: 140, propertyCount: 80, individualCount: 1100 },
    { date: '2026-03-08', classCount: 148, propertyCount: 85, individualCount: 1150 },
    { date: '2026-03-15', classCount: 156, propertyCount: 89, individualCount: 1204 },
  ],
};

const MOCK_DEPLOYMENTS = [
  { id: 'dep-001', url: 'https://example.com/onto/1', status: 'active', createdAt: '2026-02-01T10:00:00Z' },
  { id: 'dep-002', url: 'https://example.com/onto/2', status: 'stopped', createdAt: '2026-01-15T10:00:00Z' },
];

const MOCK_MERGE_REQUESTS = [
  { id: 'mr-001', title: 'Add Person class hierarchy', status: 'open', author: 'editor_seed', createdAt: '2026-03-14T10:00:00Z' },
  { id: 'mr-002', title: 'Fix property constraints', status: 'merged', author: 'owner_seed', createdAt: '2026-03-10T10:00:00Z' },
];

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
    { id: 'act-001', text: 'Committed to University Ontology', author: 'owner_seed', timestamp: '2026-03-15T14:30:00Z', type: 'commit' },
    { id: 'act-002', text: 'Created merge request MR-002', author: 'editor_seed', timestamp: '2026-03-15T12:00:00Z', type: 'merge_request' },
    { id: 'act-003', text: 'Updated class Person', author: 'viewer_seed', timestamp: '2026-03-14T16:45:00Z', type: 'class_edit' },
  ],
  recentOntologies: [
    { id: 'ont-001', name: 'University Ontology', description: 'Academic ontology', visibility: 'private', updatedAt: '2026-03-15T14:30:00Z' },
    { id: 'ont-002', name: 'Healthcare Terms', description: 'Medical terminology', visibility: 'internal', updatedAt: '2026-03-14T12:00:00Z' },
    { id: 'ont-003', name: 'Financial Taxonomy', description: 'Finance classification', visibility: 'public', updatedAt: '2026-03-13T10:00:00Z' },
  ],
};

const MOCK_VALIDATION = { status: 'ok', violations: [], validatedAt: '2026-03-15T14:30:00Z' };

// Mock data for GraphQL queries (matching query field names exactly)
const MOCK_PROJECTS = {
  items: [
    { id: 'proj-001', name: 'University Ontology', description: 'Academic ontology project', visibility: 'private', ontologyCount: 3, memberCount: 5, updatedAt: '2026-03-15T14:30:00Z' },
    { id: 'proj-002', name: 'Healthcare Terms', description: 'Medical terminology project', visibility: 'internal', ontologyCount: 2, memberCount: 8, updatedAt: '2026-03-14T12:00:00Z' },
    { id: 'proj-003', name: 'Financial Taxonomy', description: 'Finance project', visibility: 'public', ontologyCount: 1, memberCount: 3, updatedAt: '2026-03-13T10:00:00Z' },
  ],
  total: 3,
  page: 1,
  perPage: 50,
};

const MOCK_GROUPS_GQL = {
  items: [
    { id: 'grp-001', name: 'Engineering', description: 'Engineering team', parentGroupId: null, childGroups: [{ id: 'grp-002', name: 'Data Science' }], memberCount: 12, projectCount: 5 },
    { id: 'grp-002', name: 'Data Science', description: 'Data science team', parentGroupId: 'grp-001', childGroups: [], memberCount: 8, projectCount: 3 },
    { id: 'grp-003', name: 'Research', description: 'Research division', parentGroupId: null, childGroups: [], memberCount: 5, projectCount: 2 },
  ],
};

const MOCK_MEMBERS_GQL = [
  { id: 'user-456', userId: 'user-456', username: 'owner_seed', avatarUrl: '', role: 'owner', addedAt: '2026-01-15T10:00:00Z' },
  { id: 'user-789', userId: 'user-789', username: 'editor_seed', avatarUrl: '', role: 'editor', addedAt: '2026-02-20T14:30:00Z' },
  { id: 'user-012', userId: 'user-012', username: 'viewer_seed', avatarUrl: '', role: 'viewer', addedAt: '2026-03-10T09:15:00Z' },
];

const MOCK_METRICS_GQL = {
  counters: { classCount: 156, propertyCount: 89, individualCount: 1204, axiomCount: 18450, commentCount: 42, mergeRequestCount: 3 },
  trends: [
    { date: '2026-03-01', classCount: 140, propertyCount: 80, individualCount: 1100 },
    { date: '2026-03-08', classCount: 148, propertyCount: 85, individualCount: 1150 },
    { date: '2026-03-15', classCount: 156, propertyCount: 89, individualCount: 1204 },
  ],
};

const MOCK_DEPLOYMENTS_GQL = [
  { id: 'dep-001', url: 'https://example.com/onto/1', status: 'active', version: '1.0', ontologyId: 'ont-001', ontologyName: 'University Ontology', deployedAt: '2026-02-01T10:00:00Z', deployedBy: 'owner_seed' },
  { id: 'dep-002', url: 'https://example.com/onto/2', status: 'stopped', version: '2.0', ontologyId: 'ont-002', ontologyName: 'Healthcare Terms', deployedAt: '2026-01-15T10:00:00Z', deployedBy: 'editor_seed' },
];

const MOCK_MERGE_REQUESTS_GQL = [
  { id: 'mr-001', title: 'Add Person class hierarchy', description: 'Adding Person class with subclasses', sourceBranch: 'feature/person-hierarchy', targetBranch: 'main', authorName: 'editor_seed', status: 'open', mergeStatus: 'unchecked', createdAt: '2026-03-14T10:00:00Z', commentCount: 2 },
  { id: 'mr-002', title: 'Fix property constraints', description: 'Fix property range constraints', sourceBranch: 'fix/property-constraints', targetBranch: 'main', authorName: 'owner_seed', status: 'merged', mergeStatus: 'merged', createdAt: '2026-03-10T10:00:00Z', commentCount: 0 },
];

const MOCK_COMMITS = {
  items: [
    { id: 'commit-001', branchId: 'branch-main', parentCommitId: 'commit-000', message: 'Initial ontology setup', authorId: 'user-456', authorName: 'owner_seed', totalChanges: 15, createdAt: '2026-03-01T10:00:00Z' },
    { id: 'commit-002', branchId: 'branch-main', parentCommitId: 'commit-001', message: 'Add Person class', authorId: 'user-789', authorName: 'editor_seed', totalChanges: 3, createdAt: '2026-03-05T14:00:00Z' },
    { id: 'commit-003', branchId: 'branch-main', parentCommitId: 'commit-002', message: 'Add properties', authorId: 'user-789', authorName: 'editor_seed', totalChanges: 8, createdAt: '2026-03-10T09:00:00Z' },
  ],
  total: 3,
  page: 1,
  perPage: 50,
};

const MOCK_BRANCHES = {
  items: [
    { id: 'branch-main', name: 'main', ontologyId: 'ont-001', headCommitId: 'commit-003', createdAt: '2026-01-15T10:00:00Z', isProtected: true, lastCommitMessage: 'Add properties', lastCommitAuthor: 'editor_seed', aheadCount: 0, behindCount: 0 },
    { id: 'branch-dev', name: 'develop', ontologyId: 'ont-001', headCommitId: 'commit-002', createdAt: '2026-02-01T10:00:00Z', isProtected: false, lastCommitMessage: 'Add Person class', lastCommitAuthor: 'editor_seed', aheadCount: 1, behindCount: 0 },
  ],
};

const MOCK_TAGS_GQL = [
  { id: 'tag-001', name: 'v1.0.0', commitId: 'commit-003', message: 'Initial release', authorName: 'owner_seed', createdAt: '2026-03-15T10:00:00Z' },
  { id: 'tag-002', name: 'v0.9.0', commitId: 'commit-002', message: 'Beta', authorName: 'editor_seed', createdAt: '2026-03-10T10:00:00Z' },
];

const MOCK_SPARQL_RESULTS = {
  columns: ['s', 'p', 'o'],
  rows: [
    ['ex:Person', 'rdf:type', 'owl:Class'],
    ['ex:Student', 'rdfs:subClassOf', 'ex:Person'],
  ],
  total: 2,
  executionTimeMs: 15,
};

// Rate limiting state — per-endpoint tracking, only /health triggers 429
const healthRequestCounts = new Map();

function isRateLimited(path, clientIp) {
  if (!path.endsWith('/health')) return false;
  const count = (healthRequestCounts.get(clientIp) || 0) + 1;
  healthRequestCounts.set(clientIp, count);
  return count > 10;
}

function parseBody(req) {
  return new Promise((resolve) => {
    let body = '';
    req.on('data', (chunk) => { body += chunk; });
    req.on('end', () => {
      try { resolve(JSON.parse(body)); }
      catch { resolve(body); }
    });
  });
}

// Delay helper — provides realistic latency for loading-state tests
function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function sendJson(res, statusCode, data) {
  res.writeHead(statusCode, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify(data));
}

// Check auth headers and return appropriate status or null (pass)
function checkAuth(req) {
  const authHeader = req.headers['authorization'];

  // Explicitly empty auth → 401 (MUST check BEFORE !authHeader since '' is falsy)
  if (authHeader === '') return 401;

  // No auth header set → allow (normal tests without specific auth requirement)
  if (!authHeader) return null;

  // Expired token → 401
  if (authHeader === 'Bearer expired.jwt.token') return 401;

  // Viewer read-only check
  if (authHeader === 'Bearer viewer.token') {
    if (req.method === 'POST' || req.method === 'PUT' || req.method === 'DELETE') {
      return 403;
    }
  }

  return null; // Auth passed
}

async function handleRequest(req, res) {
  try {
    const url = new URL(req.url, `http://localhost:${PORT}`);
  const path = url.pathname;
  const method = req.method;

  console.log(`[stub] ${method} ${path}`);

  // Rate limiting check — only /health endpoint
  const clientIp = req.headers['x-forwarded-for'] || req.socket.remoteAddress || 'unknown';
  if (isRateLimited(path, clientIp)) {
    sendJson(res, 429, { error: 'RATE_LIMITED', message: 'Too many requests' });
    return;
  }

  // Auth check (headers only, no body needed)
  const authStatus = checkAuth(req);
  if (authStatus !== null) {
    sendJson(res, authStatus, authStatus === 401
      ? { error: 'UNAUTHORIZED', message: 'Authentication required' }
      : { error: 'FORBIDDEN', message: 'Insufficient permissions' }
    );
    return;
  }

  // Parse body ONCE for POST/PUT — reused by all route handlers and malformed JSON check
  let requestBody = null;
  if (method === 'POST' || method === 'PUT') {
    requestBody = await parseBody(req);

    // Malformed JSON check
    if (requestBody === 'not-json') {
      sendJson(res, 400, { error: 'BAD_REQUEST', message: 'Malformed JSON' });
      return;
    }
  }

  // Route matching
  if (method === 'GET' && path === '/api/v1/ontologies') {
    return sendJson(res, 200, MOCK_ONTOLOGIES);
  }

  if (method === 'GET' && path.match(/^\/api\/v1\/ontologies\/([^/]+)$/)) {
    const id = path.split('/').pop();
    const onto = MOCK_ONTOLOGIES.find(o => o.id === id);
    if (!onto) return sendJson(res, 404, { error: 'NOT_FOUND', message: 'Ontology not found' });
    return sendJson(res, 200, onto);
  }

  if (method === 'POST' && path === '/api/v1/ontologies') {
    return sendJson(res, 200, { id: 'ont-new-001', name: 'Test Ontology', visibility: 'private' });
  }

  if (method === 'GET' && path === '/api/v1/groups') {
    return sendJson(res, 200, MOCK_GROUPS);
  }

  if (method === 'GET' && path.match(/^\/api\/v1\/ontologies\/([^/]+)\/members$/)) {
    return sendJson(res, 200, MOCK_MEMBERS);
  }

  if (method === 'PUT' && path.match(/^\/api\/v1\/ontologies\/([^/]+)\/members\/([^/]+)$/)) {
    return sendJson(res, 200, { success: true });
  }

  if (method === 'DELETE' && path.match(/^\/api\/v1\/ontologies\/([^/]+)\/members\/([^/]+)$/)) {
    return sendJson(res, 200, { success: true });
  }

  if (method === 'GET' && path.match(/^\/api\/v1\/versioning\/([^/]+)\/tags$/)) {
    return sendJson(res, 200, MOCK_TAGS);
  }

  if (method === 'POST' && path.match(/^\/api\/v1\/versioning\/([^/]+)\/compare$/)) {
    return sendJson(res, 200, { changes: [], summary: { additions: 5, deletions: 3 } });
  }

  if (method === 'GET' && path.match(/^\/api\/v1\/metrics\/([^/]+)$/)) {
    return sendJson(res, 200, MOCK_METRICS);
  }

  if (method === 'POST' && path.match(/^\/api\/v1\/validation\/([^/]+)\/run$/)) {
    return sendJson(res, 200, MOCK_VALIDATION);
  }

  if (method === 'GET' && path === '/api/v1/deployments') {
    return sendJson(res, 200, MOCK_DEPLOYMENTS);
  }

  if (method === 'GET' && path === '/api/v1/merge-requests') {
    return sendJson(res, 200, MOCK_MERGE_REQUESTS);
  }

  if (method === 'GET' && (path === '/api/v1/health' || path === '/health')) {
    return sendJson(res, 200, { status: 'ok', service: 'stub-api' });
  }

  if (method === 'GET' && (path === '/api/v1/ready' || path === '/ready')) {
    return sendJson(res, 200, { status: 'ready', service: 'stub-api' });
  }

  // GraphQL endpoint
  if ((method === 'POST' && path === '/api/v1/graphql') || (method === 'POST' && path === '/graphql')) {
    const query = (requestBody?.query || '').replace(/\s+/g, ' ').trim();
    const opName = requestBody?.operationName || '';

    // DASHBOARD_QUERY: query DashboardAggregate
    if (query.includes('DashboardAggregate') || query.includes('{ dashboard ')) {
      return sendJson(res, 200, { data: { dashboard: MOCK_DASHBOARD } });
    }

    // ── Artificial delays for loading-state tests ──────────────────────────────
    // These delays ensure loading indicators render long enough for Playwright
    // assertions like toBeVisible({ timeout: 2000 }) to capture them.

    if (query.includes('OntologyMetrics') || query.includes('{ ontologyMetrics(')) {
      await delay(400);
      return sendJson(res, 200, { data: { ontologyMetrics: MOCK_METRICS_GQL } });
    }

    if (query.includes('SparqlExecute') || query.includes('{ sparqlQuery(')) {
      await delay(300);
      return sendJson(res, 200, { data: { sparqlQuery: MOCK_SPARQL_RESULTS } });
    }

    if (query.includes('RunValidation') || query.includes('{ runValidation(')) {
      await delay(300);
      return sendJson(res, 200, { data: { runValidation: MOCK_VALIDATION } });
    }

    if (query.includes('GetCommitHistory') || query.includes('{ commits(')) {
      await delay(200);
      return sendJson(res, 200, { data: { commits: MOCK_COMMITS } });
    }

    if (query.includes('GetBranches') || query.includes('{ branches(')) {
      await delay(200);
      return sendJson(res, 200, { data: { branches: MOCK_BRANCHES } });
    }

    if (query.includes('GetTags') || query.includes('{ tags(')) {
      await delay(200);
      return sendJson(res, 200, { data: { tags: MOCK_TAGS_GQL } });
    }

    if (query.includes('CompareRevisions') || query.includes('{ compareRevisions(')) {
      await delay(200);
      return sendJson(res, 200, {
        data: {
          compareRevisions: {
            additions: 5,
            deletions: 3,
            changes: [
              { entityId: 'cls-001', entityType: 'class', entityLabel: 'Person', changeType: 'added', field: 'label', oldValue: null, newValue: 'Person' },
              { entityId: 'cls-002', entityType: 'class', entityLabel: 'Student', changeType: 'added', field: 'label', oldValue: null, newValue: 'Student' },
            ],
          },
        },
      });
    }

    if (query.includes('GraphNeighborhood') || query.includes('{ graphNeighborhood(')) {
      await delay(200);
      return sendJson(res, 200, {
        data: {
          graphNeighborhood: {
            nodes: [
              { id: 'node-1', label: 'Person' },
              { id: 'node-2', label: 'Student' },
              { id: 'node-3', label: 'Professor' },
            ],
            edges: [
              { sourceId: 'node-1', targetId: 'node-2' },
              { sourceId: 'node-1', targetId: 'node-3' },
            ],
          },
        },
      });
    }

    // LIST_PROJECTS_QUERY: query ListProjects
    if (query.includes('ListProjects') || query.includes('{ projects(')) {
      return sendJson(res, 200, { data: { projects: MOCK_PROJECTS } });
    }

    // LIST_GROUPS_QUERY: query ListGroups
    if (query.includes('ListGroups') || query.includes('{ groups(')) {
      return sendJson(res, 200, { data: { groups: MOCK_GROUPS_GQL } });
    }

    // LIST_MEMBERS_QUERY: query ListMembers
    if (query.includes('ListMembers') || query.includes('{ members(')) {
      return sendJson(res, 200, { data: { members: MOCK_MEMBERS_GQL } });
    }

    // LIST_DEPLOYMENTS_QUERY: query ListDeployments
    if (query.includes('ListDeployments') || query.includes('{ deployments(')) {
      return sendJson(res, 200, { data: { deployments: MOCK_DEPLOYMENTS_GQL } });
    }

    // LIST_MERGE_REQUESTS_QUERY: query ListMergeRequests
    if (query.includes('ListMergeRequests') || query.includes('{ mergeRequests(')) {
      return sendJson(res, 200, { data: { mergeRequests: MOCK_MERGE_REQUESTS_GQL } });
    }



    // UpdateDraft mutation
    if (query.includes('UpdateDraft') || query.includes('{ updateDraft(')) {
      return sendJson(res, 200, {
        data: { updateDraft: { success: true, timestamp: new Date().toISOString() } },
      });
    }

    // Default GraphQL response
    return sendJson(res, 200, { data: {} });
  }

  // 404 for unmatched routes
    sendJson(res, 404, { error: 'NOT_FOUND', message: `Route not found: ${method} ${path}` });
  } catch (e) {
    console.error(`[stub] UNHANDLED ERROR: ${e.message}`, e.stack);
    try { sendJson(res, 500, { error: 'INTERNAL', message: e.message }); } catch (_) {}
  }
}

const server = http.createServer(handleRequest);
server.listen(PORT, '0.0.0.0', () => {
  console.log(`[stub-server] listening on http://0.0.0.0:${PORT}`);
});
