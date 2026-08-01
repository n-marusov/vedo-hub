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

// ── Error helper ──────────────────────────────────────────────────────────────────

function extractErrorMessage(err: unknown, fallback: string): string {
	if (axios.isAxiosError(err)) {
		const data = err.response?.data as Record<string, unknown> | undefined;
		return (
			(data?.error as Record<string, string> | undefined)?.message ??
			err.message ??
			fallback
		);
	}
	if (err instanceof Error) {
		return err.message;
	}
	return fallback;
}

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
	console.info(
		JSON.stringify({
			level: "info",
			msg: "org.groups.list.request",
			q,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const params = q ? `?search=${encodeURIComponent(q)}` : "";
		const { data } = await api.get(`/groups${params}`);
		console.info(
			JSON.stringify({
				level: "info",
				msg: "org.groups.list.success",
				count: data.data?.length ?? 0,
				ts: new Date().toISOString(),
			}),
		);
		// Map API snake_case fields to the camelCase GroupInfo interface.
		// The REST groups endpoint returns parent_id / childGroups not nested.
		return (data.data ?? []).map((g: Record<string, unknown>) => ({
			id: String(g.id ?? ""),
			name: String(g.name ?? ""),
			description: (g.description as string | null) ?? null,
			parentGroupId: (g.parent_id as string | null) ?? null,
			visibility: String(g.visibility ?? "Private").toLowerCase(),
			memberCount: Number(g.member_count ?? 0),
			projectCount: Number(g.project_count ?? 0),
		}));
	} catch (err: unknown) {
		const msg = extractErrorMessage(err, "Failed to list groups");
		console.error(
			JSON.stringify({
				level: "error",
				msg: "org.groups.list.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		throw new Error(msg);
	}
}

export async function createGroup(params: {
	name: string;
	slug?: string;
	description?: string;
	visibility?: string;
	parent_id?: string | null;
}): Promise<GroupInfo> {
	console.info(
		JSON.stringify({
			level: "info",
			msg: "org.groups.create.request",
			name: params.name,
			visibility: params.visibility,
			parent_id: params.parent_id,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const payload: Record<string, unknown> = {
			name: params.name,
			description: params.description,
			visibility: params.visibility,
			parent_id: params.parent_id,
		};
		if (params.slug) {
			payload.slug = params.slug;
		}

		const { data } = await api.post("/groups", payload);
		console.info(
			JSON.stringify({
				level: "info",
				msg: "org.groups.create.success",
				id: data.data?.id ?? data.id,
				name: params.name,
				ts: new Date().toISOString(),
			}),
		);
		return data.data ?? data;
	} catch (err: unknown) {
		const msg = extractErrorMessage(err, "Failed to create group");
		console.error(
			JSON.stringify({
				level: "error",
				msg: "org.groups.create.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		throw new Error(msg);
	}
}

export async function getGroup(id: string): Promise<GroupInfo> {
	console.info(
		JSON.stringify({
			level: "info",
			msg: "org.group.get.request",
			id,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const { data } = await api.get(`/groups/${id}`);
		return data.data ?? data;
	} catch (err: unknown) {
		const msg = extractErrorMessage(err, "Failed to get group");
		console.error(
			JSON.stringify({
				level: "error",
				msg: "org.group.get.failed",
				id,
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
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
	console.info(
		JSON.stringify({
			level: "info",
			msg: "org.projects.list.request",
			...params,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const searchParams = new URLSearchParams();
		if (params?.q) searchParams.set("search", params.q);
		if (params?.sortBy) searchParams.set("sort_by", params.sortBy);
		if (params?.sortDir) searchParams.set("sort_dir", params.sortDir);
		if (params?.page !== undefined)
			searchParams.set("page", String(params.page));
		if (params?.perPage !== undefined)
			searchParams.set("per_page", String(params.perPage));
		const qs = searchParams.toString();
		const { data } = await api.get(`/projects${qs ? `?${qs}` : ""}`);
		console.info(
			JSON.stringify({
				level: "info",
				msg: "org.projects.list.success",
				total: data.total ?? data.data?.length ?? 0,
				ts: new Date().toISOString(),
			}),
		);
		return { items: data.data ?? [], total: data.total ?? 0 };
	} catch (err: unknown) {
		const msg = extractErrorMessage(err, "Failed to list projects");
		console.error(
			JSON.stringify({
				level: "error",
				msg: "org.projects.list.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		throw new Error(msg);
	}
}

export async function createProject(params: {
	name: string;
	description?: string;
	groupId?: string | null;
	visibility?: string;
}): Promise<ProjectInfo> {
	console.info(
		JSON.stringify({
			level: "info",
			msg: "org.projects.create.request",
			name: params.name,
			groupId: params.groupId,
			visibility: params.visibility,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const payload: Record<string, unknown> = {
			name: params.name,
			description: params.description,
			group_id: params.groupId ?? undefined,
			visibility: params.visibility ?? "Private",
		};

		// Add Idempotency-Key for safe retry (REQ-FUN.API.write-idempotency).
		const idempotencyKey =
			crypto.randomUUID?.() ?? `${Date.now()}-${Math.random()}`;

		const { data } = await api.post("/projects", payload, {
			headers: { "Idempotency-Key": idempotencyKey },
		});
		console.info(
			JSON.stringify({
				level: "info",
				msg: "org.projects.create.success",
				id: data.data?.id ?? data.id,
				name: params.name,
				ts: new Date().toISOString(),
			}),
		);
		return data.data ?? data;
	} catch (err: unknown) {
		const msg = extractErrorMessage(err, "Failed to create project");
		console.error(
			JSON.stringify({
				level: "error",
				msg: "org.projects.create.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		throw new Error(msg);
	}
}

export async function getProject(id: string): Promise<ProjectInfo> {
	console.info(
		JSON.stringify({
			level: "info",
			msg: "org.project.get.request",
			id,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const { data } = await api.get(`/projects/${id}`);
		return data.data ?? data;
	} catch (err: unknown) {
		const msg = extractErrorMessage(err, "Failed to get project");
		console.error(
			JSON.stringify({
				level: "error",
				msg: "org.project.get.failed",
				id,
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		throw new Error(msg);
	}
}

// ── Members ────────────────────────────────────────────────────────────────────────

export async function listMembers(projectId: string): Promise<MemberInfo[]> {
	console.info(
		JSON.stringify({
			level: "info",
			msg: "org.members.list.request",
			projectId,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const { data } = await api.get(`/projects/${projectId}/members`);
		console.info(
			JSON.stringify({
				level: "info",
				msg: "org.members.list.success",
				projectId,
				count: data.data?.length ?? 0,
				ts: new Date().toISOString(),
			}),
		);
		// Map API snake_case fields to the camelCase MemberInfo interface.
		// The members endpoint returns user_id; username/avatarUrl are absent.
		return (data.data ?? []).map((m: Record<string, unknown>) => ({
			userId: String(m.user_id ?? ""),
			username: (m.username as string | null) ?? null,
			avatarUrl: (m.avatar_url as string | null) ?? null,
			role: String(m.role ?? "Viewer"),
			addedAt: (m.added_at as string | null) ?? null,
		}));
	} catch (err: unknown) {
		const msg = extractErrorMessage(err, "Failed to list members");
		console.error(
			JSON.stringify({
				level: "error",
				msg: "org.members.list.failed",
				projectId,
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		throw new Error(msg);
	}
}
