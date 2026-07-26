// @m8 — Fork API: REST client for project fork operations
// Validates: REQ-NFR.SECURITY.organization-access-model
// Uses REST (not GraphQL) per ADR-DES.API.rest-graphql-mutation-boundary
import axios from "axios";

/** ID of the VEDO Demos group in the URL path. */
export const VEDO_DEMOS_GROUP_SLUG = "vedo-demos";

export interface ForkResponse {
	project_id: string;
	ontology_id: string;
	upstream_project_id: string;
}

export interface DemoProject {
	id: string;
	name: string;
	description: string;
	classCount: number;
	propertyCount: number;
	domain: string;
}

/**
 * Fork a project by source project ID.
 * POST /api/v1/projects/{id}/fork with Idempotency-Key header.
 */
export async function forkProject(
	sourceProjectId: string,
): Promise<ForkResponse> {
	const res = await axios.post(
		`/api/v1/projects/${sourceProjectId}/fork`,
		{},
		{
			headers: { "Idempotency-Key": crypto.randomUUID() },
		},
	);
	return res.data;
}

/**
 * Fetch demo projects from the VEDO Demos group.
 * GET /api/v1/groups/{vedo-demos-slug}/projects
 */
export async function fetchDemoProjects(): Promise<DemoProject[]> {
	const res = await axios.get(
		`/api/v1/groups/${VEDO_DEMOS_GROUP_SLUG}/projects`,
	);
	return res.data.projects;
}
