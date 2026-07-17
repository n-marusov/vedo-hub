package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	ontologyv1 "vedo-core/src/services/shared/proto/ontology/v1"
)

// OntologyServiceClient is the interface for fetching ontology context.
// Implemented by the proxy.OntologyServiceClient in the Gateway.
type OntologyServiceClient interface {
	ListClasses(ctx context.Context, req *ontologyv1.ListClassesRequest) (*ontologyv1.ListClassesResponse, error)
	GetOntology(ctx context.Context, req *ontologyv1.GetOntologyRequest) (*ontologyv1.GetOntologyResponse, error)
}

// buildOntologyContext constructs a text representation of the ontology
// by fetching real class data from the ontology service via gRPC.
// Falls back to basic stub when the gRPC client is not available.
func buildOntologyContext(ctx context.Context, ontologyGrpc OntologyServiceClient, ontologyID, classID string, _ int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Ontology: %s\n", ontologyID))
	if classID != "" {
		b.WriteString(fmt.Sprintf("Focus Class: %s\n", classID))
	}

	if ontologyGrpc != nil && ontologyID != "" {
		classes, err := ontologyGrpc.ListClasses(ctx, &ontologyv1.ListClassesRequest{
			OntologyId: ontologyID,
		})
		if err != nil {
			slog.Warn("handler.build_ontology_context.list_classes_failed",
				"ontology_id", ontologyID,
				"error", err,
			)
			b.WriteString("(Unable to fetch full ontology context)\n")
			return b.String()
		}

		if classes != nil && len(classes.GetClasses()) > 0 {
			b.WriteString("Existing classes:\n")
			count := 0
			for _, cls := range classes.GetClasses() {
				if count >= 50 {
					b.WriteString(fmt.Sprintf("... and %d more\n", len(classes.GetClasses())-count))
					break
				}
				label := cls.GetLabel()
				if label == "" {
					label = cls.GetId()
				}
				parent := cls.GetParentId()
				if parent != "" {
					b.WriteString(fmt.Sprintf("  - %s (subClassOf: %s)\n", label, parent))
				} else {
					b.WriteString(fmt.Sprintf("  - %s\n", label))
				}
				count++
			}
		}
	}

	return b.String()
}

// buildClassContext constructs a class context string for property suggestions
// by fetching real class hierarchy from the ontology service via gRPC.
func buildClassContext(ctx context.Context, ontologyGrpc OntologyServiceClient, ontologyID string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Ontology: %s\n", ontologyID))

	if ontologyGrpc != nil && ontologyID != "" {
		classes, err := ontologyGrpc.ListClasses(ctx, &ontologyv1.ListClassesRequest{
			OntologyId: ontologyID,
		})
		if err != nil {
			slog.Warn("handler.build_class_context.list_classes_failed",
				"ontology_id", ontologyID,
				"error", err,
			)
			b.WriteString("Available classes: (unable to fetch)\n")
			return b.String()
		}

		if classes != nil && len(classes.GetClasses()) > 0 {
			b.WriteString("Available classes for property assignment:\n")
			for _, cls := range classes.GetClasses() {
				label := cls.GetLabel()
				if label == "" {
					label = cls.GetId()
				}
				b.WriteString(fmt.Sprintf("  - %s\n", label))
			}
		} else {
			b.WriteString("Available classes: (none yet — suggest creating new classes)\n")
		}
	} else {
		b.WriteString("Available classes for property assignment.\n")
	}

	return b.String()
}
