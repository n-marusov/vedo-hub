package cli

import (
	"strings"
	"testing"
)

// @hlv output_format_json
func TestRenderOutput_JSONErrorOutput(t *testing.T) {
	output := CliOutput{
		Status: "error",
		Error: &CliError{
			Code:    "CLI_COMMAND_NOT_SUPPORTED",
			Message: "unsupported command",
		},
	}
	rendered := RenderOutput(output, FormatJSON)
	if !strings.Contains(rendered, "CLI_COMMAND_NOT_SUPPORTED") {
		t.Fatal("JSON error output should contain error code")
	}
}

// @hlv output_format_human
func TestRenderOutput_HumanErrorOutput(t *testing.T) {
	output := CliOutput{
		Status: "error",
		Error: &CliError{
			Code:    "CLI_INVALID_FORMAT",
			Message: "unsupported format",
		},
	}
	rendered := RenderOutput(output, FormatHuman)
	if !strings.Contains(rendered, "CLI_INVALID_FORMAT") {
		t.Fatal("human error output should contain error code")
	}
}

// @hlv output_format_json
func TestRenderOutput_JSONDiagnostics(t *testing.T) {
	output := CliOutput{
		Status: "ok",
		Data: &CliData{
			Command: "status local",
			Result:  "all-required-services-reachable",
			Diagnostics: &Diagnostics{
				ComposeServicesTotal:   28,
				ComposeServicesHealthy: 28,
			},
		},
	}
	rendered := RenderOutput(output, FormatJSON)
	if !strings.Contains(rendered, "compose_services_total") {
		t.Fatal("JSON diagnostics should contain compose_services_total")
	}
}
