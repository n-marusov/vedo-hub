// @m2.5 — Mock data for Apollo mock link — Block В (Backend Pages)
// Realistic data shapes matching queries.ts GraphQL contracts

// @m2.5 — Inline type definitions matching queries.ts GraphQL contracts
// These replace @/types/graphql (module not yet created — to be generated from schema)

interface DashboardAggregateQuery {
	dashboard: {
		widgets: Array<{
			title: string;
			count: number;
			icon: string;
			route: string;
		}>;
		recentOntologies: Array<{
			id: string;
			name: string;
			description: string;
			visibility: string;
			updatedAt: string;
		}>;
		activityFeed: Array<{
			id: string;
			text: string;
			author: string;
			timestamp: string;
			type: string;
		}>;
		attentionItems: Array<{
			id: string;
			text: string;
			severity: string;
			count: number;
		}>;
	};
}

interface OntologyMetricsQuery {
	ontologyMetrics: {
		counters: {
			classCount: number;
			propertyCount: number;
			individualCount: number;
			axiomCount: number;
			commentCount: number;
			mergeRequestCount: number;
		};
		trends: Array<{
			date: string;
			classCount: number;
			propertyCount: number;
			individualCount: number;
		}>;
	};
}

interface ListDeploymentsQuery {
	deployments: Array<{
		id: string;
		url: string;
		status: string;
		version: string;
		ontologyId: string;
		ontologyName: string;
		deployedAt: string;
		deployedBy: string;
	}>;
}

interface ListMergeRequestsQuery {
	mergeRequests: Array<{
		id: string;
		title: string;
		description: string;
		sourceBranch: string;
		targetBranch: string;
		authorName: string;
		status: string;
		mergeStatus: string;
		createdAt: string;
		commentCount: number;
	}>;
}

export const MOCK_DASHBOARD_DATA: DashboardAggregateQuery = {
	dashboard: {
		widgets: [
			{
				title: "Merge Requests",
				count: 5,
				icon: "git-merge",
				route: "/dashboard/merge_requests",
			},
			{ title: "Reviews", count: 3, icon: "eye", route: "/dashboard/reviews" },
			{
				title: "Work Items",
				count: 12,
				icon: "list-todo",
				route: "/dashboard/work_items",
			},
		],
		recentOntologies: [
			{
				id: "ont-1",
				name: "Healthcare Ontology",
				description: "Medical data model",
				visibility: "private",
				updatedAt: "2026-07-18T10:00:00Z",
			},
			{
				id: "ont-2",
				name: "E-Commerce Taxonomy",
				description: "Product categories",
				visibility: "internal",
				updatedAt: "2026-07-17T15:30:00Z",
			},
			{
				id: "ont-3",
				name: "Finance Vocabulary",
				description: "Financial terms and relations",
				visibility: "public",
				updatedAt: "2026-07-16T09:00:00Z",
			},
		],
		activityFeed: [
			{
				id: "act-1",
				text: "Alice created Merge Request #42",
				author: "Alice",
				timestamp: "2026-07-18T11:00:00Z",
				type: "merge_request",
			},
			{
				id: "act-2",
				text: "Bob committed to healthcare-ontology",
				author: "Bob",
				timestamp: "2026-07-18T10:30:00Z",
				type: "commit",
			},
			{
				id: "act-3",
				text: "Carol added comment on Class:Patient",
				author: "Carol",
				timestamp: "2026-07-18T09:00:00Z",
				type: "comment",
			},
		],
		attentionItems: [
			{
				id: "att-1",
				text: "3 merge requests need your review",
				severity: "warning",
				count: 3,
			},
			{
				id: "att-2",
				text: "Validation failed for healthcare-ontology",
				severity: "error",
				count: 1,
			},
			{
				id: "att-3",
				text: "New version available for Finance Vocabulary",
				severity: "info",
				count: 1,
			},
		],
	},
};

export const MOCK_METRICS_DATA: OntologyMetricsQuery = {
	ontologyMetrics: {
		counters: {
			classCount: 142,
			propertyCount: 89,
			individualCount: 567,
			axiomCount: 12034,
			commentCount: 45,
			mergeRequestCount: 8,
		},
		trends: [
			{
				date: "2026-07-01",
				classCount: 120,
				propertyCount: 75,
				individualCount: 480,
			},
			{
				date: "2026-07-08",
				classCount: 130,
				propertyCount: 82,
				individualCount: 510,
			},
			{
				date: "2026-07-15",
				classCount: 142,
				propertyCount: 89,
				individualCount: 567,
			},
		],
	},
};

export const MOCK_DEPLOYMENTS_DATA: ListDeploymentsQuery = {
	deployments: [
		{
			id: "dep-1",
			url: "https://healthcare.vocab.example.com",
			status: "active",
			version: "v2.1.0",
			ontologyId: "ont-1",
			ontologyName: "Healthcare Ontology",
			deployedAt: "2026-07-15T10:00:00Z",
			deployedBy: "Alice",
		},
		{
			id: "dep-2",
			url: "https://ecommerce.vocab.example.com",
			status: "active",
			version: "v1.3.0",
			ontologyId: "ont-2",
			ontologyName: "E-Commerce Taxonomy",
			deployedAt: "2026-07-10T14:00:00Z",
			deployedBy: "Bob",
		},
		{
			id: "dep-3",
			url: "https://finance.vocab.example.com",
			status: "stopped",
			version: "v0.9.0",
			ontologyId: "ont-3",
			ontologyName: "Finance Vocabulary",
			deployedAt: "2026-06-20T08:00:00Z",
			deployedBy: "Carol",
		},
	],
};

export const MOCK_MERGE_REQUESTS_DATA: ListMergeRequestsQuery = {
	mergeRequests: [
		{
			id: "mr-1",
			title: "Add Patient class properties",
			description: "Extends Patient with diagnosis and treatment fields",
			sourceBranch: "feature/patient-props",
			targetBranch: "main",
			authorName: "Alice",
			status: "open",
			mergeStatus: "can_merge",
			createdAt: "2026-07-17T10:00:00Z",
			commentCount: 3,
		},
		{
			id: "mr-2",
			title: "Refactor Healthcare Ontology structure",
			description: "Reorganizes top-level classes",
			sourceBranch: "feature/refactor",
			targetBranch: "main",
			authorName: "Bob",
			status: "open",
			mergeStatus: "conflict",
			createdAt: "2026-07-16T14:00:00Z",
			commentCount: 7,
		},
		{
			id: "mr-3",
			title: "Add E-Commerce shipping properties",
			description: "Adds dimensions, weight, shipping cost",
			sourceBranch: "feature/shipping",
			targetBranch: "main",
			authorName: "Carol",
			status: "merged",
			mergeStatus: "merged",
			createdAt: "2026-07-14T09:00:00Z",
			commentCount: 5,
		},
	],
};

// @m2.5 — Helper: create a 200ms artificial delay to simulate network latency
export function delay(ms = 200): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}
