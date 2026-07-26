// REST API client for deployment listings (publisher snapshots).
//
// After GraphQL tightening: deployments migrated from GraphQL
// (LIST_DEPLOYMENTS_QUERY) to REST.

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

export interface DeploymentInfo {
  id: string;
  url: string;
  status: string;
  version: string;
  ontologyId: string;
  ontologyName: string;
  deployedAt: string;
  deployedBy: string;
}

// ── List Deployments ──────────────────────────────────────────────────────────────

// TODO: Replace mock with real backend endpoint when implemented.
// Intended: GET /api/v1/deployments (adapter over publisher snapshots).
export async function listDeployments(includeStopped?: boolean): Promise<DeploymentInfo[]> {
  console.info(JSON.stringify({
    level: "info", msg: "deployments.list.request",
    includeStopped, ts: new Date().toISOString(),
  }));

  // Return mock data matching previous GraphQL LIST_DEPLOYMENTS_QUERY shape.
  const mock: DeploymentInfo[] = [
    {
      id: "dep-1",
      url: "https://hub.vedo.example.com/ontologies/ont-1",
      status: "active",
      version: "a1b2c3d",
      ontologyId: "ont-1",
      ontologyName: "Enterprise Architecture Model",
      deployedAt: new Date(Date.now() - 86400000).toISOString(),
      deployedBy: "alice",
    },
  ];

  const filtered = includeStopped
    ? mock
    : mock.filter((d) => d.status !== "stopped");

  console.info(JSON.stringify({
    level: "info", msg: "deployments.list.success",
    count: filtered.length, ts: new Date().toISOString(),
  }));

  return filtered;
}
