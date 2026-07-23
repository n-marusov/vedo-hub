// REST API client for organization model (groups, projects, members).
//
// After GraphQL tightening: org reads migrated from GraphQL
// (groups, projects, members) to REST endpoints in routes.go.

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

export interface GroupInfo {
  id: string;
  name: string;
  description: string | null;
  parentGroupId: string | null;
  childGroups?: GroupInfo[];
  visibility: string;
  memberCount: number;
  projectCount: number;
}

export interface ProjectInfo {
  id: string;
  name: string;
  description: string | null;
  visibility: string;
  ontologyId: string;
  memberCount: number;
  updatedAt: string | null;
}

export interface MemberInfo {
  userId: string;
  username: string | null;
  avatarUrl: string | null;
  role: string;
  addedAt: string | null;
}

// ── Groups ─────────────────────────────────────────────────────────────────────────

export async function listGroups(q?: string): Promise<GroupInfo[]> {
  console.info(JSON.stringify({
    level: "info", msg: "org.groups.list.request",
    q, ts: new Date().toISOString(),
  }));

  try {
    const params = q ? `?search=${encodeURIComponent(q)}` : "";
    const { data } = await api.get(`/groups${params}`);
    console.info(JSON.stringify({
      level: "info", msg: "org.groups.list.success",
      count: data.data?.length ?? 0, ts: new Date().toISOString(),
    }));
    return data.data ?? [];
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? "Failed to list groups";
    console.error(JSON.stringify({
      level: "error", msg: "org.groups.list.failed",
      error: msg, ts: new Date().toISOString(),
    }));
    throw new Error(msg);
  }
}

export async function createGroup(params: {
	name: string;
	description?: string;
	visibility?: string;
	parentGroupId?: string | null;
}): Promise<GroupInfo> {
	console.info(JSON.stringify({
		level: "info", msg: "org.groups.create.request",
		name: params.name, visibility: params.visibility,
		parentGroupId: params.parentGroupId, ts: new Date().toISOString(),
	}));

	try {
		const { data } = await api.post("/groups", params);
		console.info(JSON.stringify({
			level: "info", msg: "org.groups.create.success",
			id: data.data?.id ?? data.id,
			name: params.name, ts: new Date().toISOString(),
		}));
		return data.data ?? data;
	} catch (err: any) {
		const msg = err.response?.data?.error?.message ?? err.message ?? "Failed to create group";
		console.error(JSON.stringify({
			level: "error", msg: "org.groups.create.failed",
			error: msg, ts: new Date().toISOString(),
		}));
		throw new Error(msg);
	}
}

export async function getGroup(id: string): Promise<GroupInfo> {
  console.info(JSON.stringify({
    level: "info", msg: "org.group.get.request",
    id, ts: new Date().toISOString(),
  }));

  try {
    const { data } = await api.get(`/groups/${id}`);
    return data.data ?? data;
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? "Failed to get group";
    console.error(JSON.stringify({
      level: "error", msg: "org.group.get.failed",
      id, error: msg, ts: new Date().toISOString(),
    }));
    throw new Error(msg);
  }
}

// ── Projects ───────────────────────────────────────────────────────────────────────

export async function listProjects(params?: {
  q?: string;
  sortBy?: string;
  sortDir?: string;
  page?: number;
  perPage?: number;
}): Promise<{ items: ProjectInfo[]; total: number }> {
  console.info(JSON.stringify({
    level: "info", msg: "org.projects.list.request",
    ...params, ts: new Date().toISOString(),
  }));

  try {
    const searchParams = new URLSearchParams();
    if (params?.q) searchParams.set("search", params.q);
    if (params?.sortBy) searchParams.set("sort_by", params.sortBy);
    if (params?.sortDir) searchParams.set("sort_dir", params.sortDir);
    if (params?.page !== undefined) searchParams.set("page", String(params.page));
    if (params?.perPage !== undefined) searchParams.set("per_page", String(params.perPage));
    const qs = searchParams.toString();
    const { data } = await api.get(`/projects${qs ? "?" + qs : ""}`);
    console.info(JSON.stringify({
      level: "info", msg: "org.projects.list.success",
      total: data.total ?? data.data?.length ?? 0, ts: new Date().toISOString(),
    }));
    return { items: data.data ?? [], total: data.total ?? 0 };
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? "Failed to list projects";
    console.error(JSON.stringify({
      level: "error", msg: "org.projects.list.failed",
      error: msg, ts: new Date().toISOString(),
    }));
    throw new Error(msg);
  }
}

export async function getProject(id: string): Promise<ProjectInfo> {
  console.info(JSON.stringify({
    level: "info", msg: "org.project.get.request",
    id, ts: new Date().toISOString(),
  }));

  try {
    const { data } = await api.get(`/projects/${id}`);
    return data.data ?? data;
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? "Failed to get project";
    console.error(JSON.stringify({
      level: "error", msg: "org.project.get.failed",
      id, error: msg, ts: new Date().toISOString(),
    }));
    throw new Error(msg);
  }
}

// ── Members ────────────────────────────────────────────────────────────────────────

export async function listMembers(projectId: string): Promise<MemberInfo[]> {
  console.info(JSON.stringify({
    level: "info", msg: "org.members.list.request",
    projectId, ts: new Date().toISOString(),
  }));

  try {
    const { data } = await api.get(`/projects/${projectId}/members`);
    console.info(JSON.stringify({
      level: "info", msg: "org.members.list.success",
      projectId, count: data.data?.length ?? 0, ts: new Date().toISOString(),
    }));
    return data.data ?? [];
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? "Failed to list members";
    console.error(JSON.stringify({
      level: "error", msg: "org.members.list.failed",
      projectId, error: msg, ts: new Date().toISOString(),
    }));
    throw new Error(msg);
  }
}
