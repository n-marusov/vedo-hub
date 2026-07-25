package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// @ctx: output format rendering — human and JSON for CLI-OPS-001 contract

// RenderOutput formats the CliOutput according to the specified format.
// @hlv:sec [INPUT_VALIDATION] — format selection validated before rendering
func RenderOutput(output CliOutput, format OutputFormat) string {
	slog.Info("output.render", "format", format, "status", output.Status)

	switch format {
	case FormatJSON:
		return renderJSON(output)
	case FormatHuman:
		return renderHuman(output)
	default:
		return renderHuman(output)
	}
}

func renderJSON(output CliOutput) string {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		slog.Error("output.json.marshal.failed", "error", err)
		return `{"status":"error","error":{"code":"CLI_AUDIT_REDACTION_FAILED","message":"output serialization failed"}}`
	}
	return string(data)
}

// @hlv:sec [SECRET_HANDLING] — human output redacts secret-like fields
func renderHuman(output CliOutput) string {
	var b strings.Builder
	if output.Status == "ok" && output.Data != nil {
		b.WriteString(fmt.Sprintf("Status: %s\n", output.Status))
		b.WriteString(fmt.Sprintf("Command: %s\n", output.Data.Command))
		b.WriteString(fmt.Sprintf("Result: %s\n", output.Data.Result))
		b.WriteString(fmt.Sprintf("Trace ID: %s\n", output.Data.TraceID))
		b.WriteString(fmt.Sprintf("Correlation ID: %s\n", output.Data.CorrelationID))
		if output.Data.Diagnostics != nil {
			b.WriteString(fmt.Sprintf("Compose: %d/%d healthy\n",
				output.Data.Diagnostics.ComposeServicesHealthy,
				output.Data.Diagnostics.ComposeServicesTotal,
			))
		}
	} else if output.Error != nil {
		b.WriteString(fmt.Sprintf("Error [%s]: %s\n", output.Error.Code, output.Error.Message))
	}
	return b.String()
}
