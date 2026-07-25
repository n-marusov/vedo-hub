package cli

// @ctx: type definitions for CLI-OPS-001 contract — command, output format, result

type Command string

const (
	CmdSupportTenantInfo            Command = "support tenant-info"
	CmdSupportAuditTrail            Command = "support audit-trail"
	CmdSupportListBackups           Command = "support list-backups"
	CmdEmergencyReadonly            Command = "emergency readonly"
	CmdEmergencyClear               Command = "emergency clear"
	CmdSupportEmergencyAccessReq    Command = "support emergency-access request"
	CmdSupportEmergencyAccessList   Command = "support emergency-access list"
	CmdSupportEmergencyAccessRevoke Command = "support emergency-access revoke"
	CmdSecurityPolicyDisable        Command = "security policy disable"

	CmdStatusLocal      Command = "status local"
	CmdDiagnoseCompose  Command = "diagnose compose"
	CmdDocsOpen         Command = "docs open"
	CmdAuthResolveCreds Command = "auth resolve-credentials"

	CmdTicketCreate  Command = "ticket create"
	CmdTicketList    Command = "ticket list"
	CmdTicketShow    Command = "ticket show"
	CmdTicketComment Command = "ticket comment"
	CmdTicketUpdate  Command = "ticket update"
	CmdTicketClose   Command = "ticket close"
	CmdTicketReopen  Command = "ticket reopen"
	CmdTicketDelete  Command = "ticket delete"
)

type OutputFormat string

const (
	FormatHuman OutputFormat = "human"
	FormatJSON  OutputFormat = "json"
)

type TargetEnvironment string

const (
	EnvLocal     TargetEnvironment = "local"
	EnvCI        TargetEnvironment = "ci"
	EnvAirgapped TargetEnvironment = "airgapped"
)

type GuardrailLevel string

const (
	GuardrailG1 GuardrailLevel = "G1"
	GuardrailG2 GuardrailLevel = "G2"
	GuardrailG3 GuardrailLevel = "G3"
	GuardrailG4 GuardrailLevel = "G4"
)

type ActorType string

const (
	ActorUser           ActorType = "user"
	ActorServiceAccount ActorType = "service_account"
)

type CredentialSources struct {
	VaultAddr string `json:"vault_addr,omitempty"`
	AWSRegion string `json:"aws_region,omitempty"`
	EnvPrefix string `json:"env_prefix,omitempty"`
}

type CliInput struct {
	Command           Command            `json:"command"`
	OutputFormat      OutputFormat       `json:"output_format"`
	TargetEnvironment TargetEnvironment  `json:"target_environment,omitempty"`
	Credentials       *CredentialSources `json:"credentials,omitempty"`
	TenantID          string             `json:"tenant_id,omitempty"`
	KeyID             string             `json:"key_id,omitempty"`
	TicketID          string             `json:"ticket_id,omitempty"`
	Reason            string             `json:"reason,omitempty"`
	GuardrailLevel    GuardrailLevel     `json:"guardrail_level,omitempty"`
	ActorType         ActorType          `json:"actor_type,omitempty"`
	MFAChallengeID    string             `json:"mfa_challenge_id,omitempty"`
	RequestID         string             `json:"request_id,omitempty"`
	IncidentID        string             `json:"incident_id,omitempty"`
	TriggerSignal     string             `json:"trigger_signal,omitempty"`
	Approvers         []string           `json:"approvers,omitempty"`
	SecretCategories  []string           `json:"secret_categories,omitempty"`
}

type CliOutput struct {
	Status string    `json:"status"` // ok, error
	Data   *CliData  `json:"data,omitempty"`
	Error  *CliError `json:"error,omitempty"`
}

type CliData struct {
	Command               string       `json:"command"`
	Result                string       `json:"result"`
	TraceID               string       `json:"trace_id"`
	CorrelationID         string       `json:"correlation_id"`
	GuardrailsApplied     []string     `json:"guardrails_applied,omitempty"`
	MFAChallengePerformed bool         `json:"mfa_challenge_performed,omitempty"`
	Diagnostics           *Diagnostics `json:"diagnostics,omitempty"`
}

type Diagnostics struct {
	ComposeServicesTotal   int `json:"compose_services_total"`
	ComposeServicesHealthy int `json:"compose_services_healthy"`
}

type CliError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
