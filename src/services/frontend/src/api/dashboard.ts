// REST API client for dashboard aggregate data.
//
// After GraphQL tightening: dashboard data migrated from GraphQL
// (DASHBOARD_QUERY) to REST. Currently returns mock data matching the
// GraphQL shape until a backend aggregator endpoint is implemented.

import axios from "axios";

const BASE = "/api/v1";

const api = axios.create({
  baseURL: BASE,
  headers: { "X-Requested-With": "XMLHttpRequest" },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("vedo-jwt-token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ── Types ──────────────────────────────────────────────────────────────────────────

export interface DashboardWidget {
  title: string;
  count: number;
  icon: string;
  route: string;
}

export interface RecentOntology {
  id: string;
  name: string;
  description: string | null;
  visibility: string;
  updatedAt: string;
}

export interface ActivityItem {
  id: string;
  text: string;
  author: string;
  timestamp: string;
  type: string;
}

export interface AttentionItem {
  id: string;
  text: string;
  severity: string;
  count: number;
}

export interface DashboardData {
  widgets: DashboardWidget[];
  recentOntologies: RecentOntology[];
  activityFeed: ActivityItem[];
  attentionItems: AttentionItem[];
}

// ── Get Dashboard ──────────────────────────────────────────────────────────────────

// TODO: Replace mock with real backend endpoint when implemented.
// Intended endpoint: GET /api/v1/dashboard
export async function getDashboard(): Promise<DashboardData> {
  console.info(JSON.stringify({
    level: "info", msg: "dashboard.get.request",
    ts: new Date().toISOString(),
  }));

  // Return mock data matching previous GraphQL DASHBOARD_QUERY shape.
  const mock: DashboardData = {
    widgets: [
      { title: "Ontologies", count: 12, icon: "database", route: "/projects" },
      { title: "Classes", count: 1543, icon: "box", route: "/workspace" },
      { title: "Properties", count: 892, icon: "link", route: "/workspace" },
      { title: "Members", count: 47, icon: "users", route: "/members" },
    ],
    recentOntologies: [
      {
        id: "ont-1",
        name: "Enterprise Architecture Model",
        description: "Core EA ontology for the organization",
        visibility: "Internal",
        updatedAt: new Date().toISOString(),
      },
    ],
    activityFeed: [
      {
        id: "act-1",
        text: "New class 'Application' created by alice",
        author: "alice",
        timestamp: new Date().toISOString(),
        type: "class_created",
      },
      {
        id: "act-2",
        text: "Commit 'Add compliance module' pushed by bob",
        author: "bob",
        timestamp: new Date(Date.now() - 3600000).toISOString(),
        type: "commit_pushed",
      },
    ],
    attentionItems: [
      { id: "att-1", text: "3 pending merge requests", severity: "warning", count: 3 },
      { id: "att-2", text: "12 classes without descriptions", severity: "info", count: 12 },
    ],
  };

  console.info(JSON.stringify({
    level: "info", msg: "dashboard.get.success",
    widgets: mock.widgets.length, ts: new Date().toISOString(),
  }));

  return mock;
}
