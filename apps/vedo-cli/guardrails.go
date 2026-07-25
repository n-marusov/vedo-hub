package cli

import (
	"fmt"
	"log/slog"
	"sync"
)

// @ctx: guardrail registry and MFA freshness hooks for CLI-OPS-001

var consumedMFAChallenges = struct {
	sync.Mutex
	ids map[string]struct{}
}{ids: map[string]struct{}{}}

var commandGuardrailLevel = map[Command]GuardrailLevel{
	CmdSupportTenantInfo:            GuardrailG1,
	CmdSupportAuditTrail:            GuardrailG2,
	CmdSupportListBackups:           GuardrailG2,
	CmdEmergencyReadonly:            GuardrailG3,
	CmdEmergencyClear:               GuardrailG4,
	CmdSupportEmergencyAccessReq:    GuardrailG3,
	CmdSupportEmergencyAccessList:   GuardrailG3,
	CmdSupportEmergencyAccessRevoke: GuardrailG4,
	CmdSecurityPolicyDisable:        GuardrailG4,

	CmdTicketCreate:  GuardrailG1,
	CmdTicketList:    GuardrailG1,
	CmdTicketShow:    GuardrailG1,
	CmdTicketComment: GuardrailG1,
	CmdTicketUpdate:  GuardrailG1,
	CmdTicketClose:   GuardrailG2,
	CmdTicketReopen:  GuardrailG1,
	CmdTicketDelete:  GuardrailG3,
}

var guardrailControls = map[GuardrailLevel][]string{
	GuardrailG1: {"typed_confirmation"},
	GuardrailG2: {"typed_confirmation", "env_guard"},
	GuardrailG3: {"typed_confirmation", "env_guard", "mfa"},
	GuardrailG4: {"typed_confirmation", "env_guard", "mfa", "cooldown_delay"},
}

// ResolveGuardrailControls returns the deterministic control set for command/level.
func ResolveGuardrailControls(command Command, requested GuardrailLevel) ([]string, GuardrailLevel, error) {
	baseLevel, ok := commandGuardrailLevel[command]
	if !ok {
		return nil, "", fmt.Errorf("unknown command guardrail level")
	}
	level := baseLevel
	if requested != "" {
		level = requested
	}
	controls, ok := guardrailControls[level]
	if !ok {
		return nil, "", fmt.Errorf("unknown guardrail level")
	}
	copyControls := make([]string, 0, len(controls))
	copyControls = append(copyControls, controls...)
	return copyControls, level, nil
}

func isCategoryAB(command Command) bool {
	return command == CmdSupportEmergencyAccessReq ||
		command == CmdEmergencyReadonly ||
		command == CmdEmergencyClear ||
		command == CmdSupportEmergencyAccessList ||
		command == CmdSupportEmergencyAccessRevoke ||
		command == CmdSecurityPolicyDisable
}

// @hlv:sec [AUTH_BOUNDARY] — service account restricted from category A/B operations
// @hlv:sec [INPUT_VALIDATION] — fresh MFA challenge required for every category A/B invocation
func evaluateMFAGuard(input CliInput) error {
	if !isCategoryAB(input.Command) {
		return nil
	}
	if input.ActorType == ActorServiceAccount {
		slog.Error("cli.guardrail.service_account_blocked",
			"request_id", input.RequestID,
			"entity_id", input.TenantID,
			"command", input.Command,
		)
		return fmt.Errorf("CLI_GUARDRAIL_NOT_SATISFIED")
	}
	if input.MFAChallengeID == "" {
		slog.Error("cli.guardrail.mfa_missing",
			"request_id", input.RequestID,
			"entity_id", input.TenantID,
			"command", input.Command,
		)
		return fmt.Errorf("CLI_MFA_REQUIRED")
	}

	consumedMFAChallenges.Lock()
	defer consumedMFAChallenges.Unlock()
	if _, used := consumedMFAChallenges.ids[input.MFAChallengeID]; used {
		slog.Error("cli.guardrail.mfa_reused",
			"request_id", input.RequestID,
			"entity_id", input.TenantID,
			"command", input.Command,
		)
		return fmt.Errorf("CLI_MFA_REQUIRED")
	}
	consumedMFAChallenges.ids[input.MFAChallengeID] = struct{}{}
	return nil
}

func resetMFAChallengesForTests() {
	consumedMFAChallenges.Lock()
	defer consumedMFAChallenges.Unlock()
	consumedMFAChallenges.ids = map[string]struct{}{}
}
