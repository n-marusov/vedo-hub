package main

import (
	"encoding/json"
	"log/slog"
)

// requiredOpenAPIKeys are the top-level keys that must be present in the
// embedded OpenAPI specification for it to be considered structurally valid.
var requiredOpenAPIKeys = []string{"openapi", "paths", "components"}

// validateOpenAPISpec parses the embedded openAPISpec and logs an ERROR-level
// message if it is not valid JSON or is missing required top-level keys.
// This runs at startup so operators catch spec corruption before clients do.
func validateOpenAPISpec() {
	var doc map[string]any
	if err := json.Unmarshal(openAPISpec, &doc); err != nil {
		slog.Error("openapi.spec.invalid",
			"error", err,
			"hint", "check apps/services/api-gateway/docs/openapi.json for JSON syntax errors",
		)
		return
	}

	for _, key := range requiredOpenAPIKeys {
		if _, ok := doc[key]; !ok {
			slog.Error("openapi.spec.missing_key",
				"key", key,
				"hint", "the embedded OpenAPI spec is missing a required top-level key",
			)
		}
	}

	paths, _ := doc["paths"].(map[string]any)
	slog.Info("openapi.spec.valid",
		"paths", len(paths),
	)
}
