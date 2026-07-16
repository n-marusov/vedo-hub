// Composable for fetching public ontology data from the public-browse-api.
//
// Provides reactive state for the public ontology view: class tree,
// property list, and metadata. Uses the ontology-service GraphQL endpoint
// (via the API Gateway) for unauthenticated public browsing.

import { gql } from "@apollo/client/core";
import { useQuery } from "@vue/apollo-composable";
import { ref, watch } from "vue";

// ── GraphQL Queries ─────────────────────────────────────────────────────────

const PUBLIC_ONTOLOGY_METADATA_QUERY = gql`
  query PublicOntologyMetadata($slug: String!) {
    publicOntology(slug: $slug) {
      id
      name
      description
      version
      publishedAt
      classCount
      propertyCount
      individualCount
    }
  }
`;

const PUBLIC_CLASS_TREE_QUERY = gql`
  query PublicClassTree($ontologyId: ID!, $maxDepth: Int) {
    classTree(ontologyId: $ontologyId, maxDepth: $maxDepth) {
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

const PUBLIC_PROPERTIES_QUERY = gql`
  query PublicProperties($ontologyId: ID!) {
    properties(ontologyId: $ontologyId, perPage: 100) {
      items {
        id
        label
        propertyType
        xsdType
      }
    }
  }
`;

// ── Types ───────────────────────────────────────────────────────────────────

export interface PublicOntology {
	id: string;
	name: string;
	description?: string;
	version: string;
	publishedAt: string;
	classCount: number;
	propertyCount: number;
	individualCount: number;
}

export interface PublicClassNode {
	id: string;
	label: string;
	children?: PublicClassNode[];
}

export interface PublicProperty {
	id: string;
	label: string;
	propertyType: string;
	xsdType?: string;
}

// ── Composable ──────────────────────────────────────────────────────────────

/// Reactive composable for loading a published ontology by slug.
///
/// Returns reactive refs for metadata, class tree, and properties.
/// All useQuery calls are at the top level (Vue setup-safe).
/// Tree and properties queries are paused until ontologyId is set.
export function usePublicOntology(slug: string) {
	const loading = ref(true);
	const error = ref<string | null>(null);
	const metadata = ref<PublicOntology | null>(null);
	const classTree = ref<PublicClassNode[]>([]);
	const properties = ref<PublicProperty[]>([]);
	const selectedClassId = ref<string | null>(null);
	const ontologyId = ref("");

	// Step 1: Fetch metadata by slug
	const metaQuery = useQuery(PUBLIC_ONTOLOGY_METADATA_QUERY, { slug });

	watch(metaQuery.result, (result) => {
		if (!result?.publicOntology) return;
		const data = result.publicOntology;
		metadata.value = {
			id: data.id,
			name: data.name,
			description: data.description,
			version: data.version,
			publishedAt: data.publishedAt,
			classCount: data.classCount,
			propertyCount: data.propertyCount,
			individualCount: data.individualCount,
		};
		ontologyId.value = data.id;
		loading.value = false;
	});

	watch(metaQuery.error, (err) => {
		if (err) {
			error.value = err.message;
			loading.value = false;
		}
	});

	// Step 2: Fetch class tree (paused until ontologyId is non-empty)
	const treeQuery = useQuery(PUBLIC_CLASS_TREE_QUERY, () => ({
		ontologyId: ontologyId.value,
		maxDepth: 3,
	}));

	watch(treeQuery.result, (result) => {
		if (result?.classTree) {
			classTree.value = result.classTree;
		}
	});

	// Step 3: Fetch properties
	const propsQuery = useQuery(PUBLIC_PROPERTIES_QUERY, () => ({
		ontologyId: ontologyId.value,
	}));

	watch(propsQuery.result, (result) => {
		if (result?.properties?.items) {
			properties.value = result.properties.items;
		}
	});

	function selectClass(id: string) {
		selectedClassId.value = id;
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
