// REST API client for merge requests.
//
// After GraphQL tightening: merge requests migrated from GraphQL
// (LIST_MERGE_REQUESTS_QUERY) to REST.

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

export interface MergeRequestInfo {
  id: string;
  title: string;
  description: string | null;
  sourceBranch: string;
  targetBranch: string;
  authorName: string;
  status: string;
  mergeStatus: string;
  createdAt: string;
  commentCount: number;
}

// ── List Merge Requests ───────────────────────────────────────────────────────────

// TODO: Replace mock with real backend endpoint when implemented.
// Intended: GET /api/v1/merge-requests (likely versioning-service).
export async function listMergeRequests(status?: string): Promise<MergeRequestInfo[]> {
  console.info(JSON.stringify({
    level: "info", msg: "mergerequests.list.request",
    status, ts: new Date().toISOString(),
  }));

  // Return mock data matching previous GraphQL LIST_MERGE_REQUESTS_QUERY shape.
  const mock: MergeRequestInfo[] = [
    {
      id: "mr-1",
      title: "Add compliance module classes",
      description: "Introduces Regulation, Policy, and Control classes with datatype properties",
      sourceBranch: "feature/compliance-module",
      targetBranch: "main",
      authorName: "bob",
      status: "open",
      mergeStatus: "can_be_merged",
      createdAt: new Date(Date.now() - 7200000).toISOString(),
      commentCount: 3,
    },
    {
      id: "mr-2",
      title: "Update Application taxonomy",
      description: null,
      sourceBranch: "fix/app-taxonomy",
      targetBranch: "main",
      authorName: "alice",
      status: "merged",
      mergeStatus: "merged",
      createdAt: new Date(Date.now() - 172800000).toISOString(),
      commentCount: 5,
    },
  ];

  const filtered = status ? mock.filter((mr) => mr.status === status) : mock;

  console.info(JSON.stringify({
    level: "info", msg: "mergerequests.list.success",
    count: filtered.length, ts: new Date().toISOString(),
  }));

  return filtered;
}
