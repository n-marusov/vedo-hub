// GraphQL Codegen configuration (client preset).
//
// Sources the canonical schema contract committed by the ontology-service
// (G1: apps/services/ontology-service/schema.graphql) and the frontend's gql
// documents, and generates fully typed document nodes + TS types.
//
// Commands:
//   pnpm codegen        — (re)generate src/types/graphql.ts
//   pnpm codegen:check  — verify generated types are up to date (CI gate)
import type { CodegenConfig } from "@graphql-codegen/cli";

const config: CodegenConfig = {
	// Canonical SDL contract exported by build_schema().sdl() (G1).
	schema: "../ontology-service/schema.graphql",
	// All frontend gql documents (.ts + .vue SFC script blocks).
	// Exclude the generated output so codegen:check doesn't re-scan it.
	documents: ["src/**/*.{ts,vue}", "!src/types/gql/**"],
	ignoreNoDocuments: true,
	generates: {
		"src/types/gql/": {
			preset: "client",
			config: {
				// Apollo Client compatible: typed document nodes + __typename.
				skipTypename: false,
				scalars: {
					ID: "string",
					DateTime: "string",
				},
			},
		},
	},
};

export default config;
