// Composable for fetching public ontology data from the public-browse-api.
//
// Provides reactive state for the public ontology view: class tree, metadata.
// Uses the public-browse-api REST surface (GET /api/v1/ontologies/{id}),
// which serves published snapshots without authentication — NOT the
// ontology-service GraphQL endpoint (the `publicOntology` field never existed
// in the graph-navigation-only schema; see ADR-DES.API.graphql-sparql-split-strategy).
//
// Contract (public-browse-api src/models.rs OntologyDetail):
//   GET /api/v1/ontologies/{id}
//   → { id, name, description, class_count, property_count, individual_count,
//       published_at, format, class_tree }
//
// Properties are not served by public-browse-api yet; the page renders the
// empty state until a public properties endpoint exists.

import axios from "axios";
import { ref } from "vue";

const PUBLIC_BASE = "/api/v1";

const api = axios.create({
	baseURL: PUBLIC_BASE,
	headers: { "X-Requested-With": "XMLHttpRequest" },
});

// ── Types (public-browse-api REST contract) ────────────────────────────────

export interface PublicClassNode {
	id: string;
	label: string;
	children?: PublicClassNode[];
}

export interface PublicOntology {
	id: string;
	name: string;
	description?: string;
	version?: string;
	publishedAt: string;
	classCount: number;
	propertyCount: number;
	individualCount: number;
}

export interface PublicProperty {
	id: string;
	label: string;
	propertyType: string;
	xsdType?: string;
}

// Raw wire shape from public-browse-api (snake_case, no class tree mapping yet).
interface PublicOntologyDetail {
	id: string;
	name: string;
	description?: string | null;
	class_count: number;
	property_count: number;
	individual_count: number;
	published_at: string;
	format: string;
	class_tree: PublicClassNode[];
}

// ── Composable ──────────────────────────────────────────────────────────────

/// Reactive composable for loading a published ontology by id/slug.
///
/// Returns reactive refs for metadata, class tree, and properties.
/// Properties stay empty until public-browse-api serves them (post-MVP).
export function usePublicOntology(id: string) {
	const loading = ref(true);
	const error = ref<string | null>(null);
	const metadata = ref<PublicOntology | null>(null);
	const classTree = ref<PublicClassNode[]>([]);
	const properties = ref<PublicProperty[]>([]);
	const selectedClassId = ref<string | null>(null);

	async function load(): Promise<void> {
		loading.value = true;
		error.value = null;
		try {
			const response = await api.get<PublicOntologyDetail>(
				`/ontologies/${encodeURIComponent(id)}`,
			);
			const detail = response.data;
			metadata.value = {
				id: detail.id,
				name: detail.name,
				description: detail.description ?? undefined,
				publishedAt: detail.published_at,
				classCount: detail.class_count,
				propertyCount: detail.property_count,
				individualCount: detail.individual_count,
			};
			classTree.value = detail.class_tree ?? [];
			loading.value = false;
		} catch (e) {
			const msg = e instanceof Error ? e.message : String(e);
			error.value = msg;
			loading.value = false;
		}
	}

	void load();

	function selectClass(nodeId: string) {
		selectedClassId.value = nodeId;
	}

	return {
		metadata,
		classTree,
		properties,
		loading,
		error,
		selectedClassId,
		selectClass,
	};
}
