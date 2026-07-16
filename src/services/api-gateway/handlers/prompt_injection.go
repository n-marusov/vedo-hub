package handlers

import (
	"errors"
	"regexp"
	"strings"
)

// Common prompt injection patterns to detect and sanitize.
var (
	// System override attempts
	reIgnoreInstructions = regexp.MustCompile(`(?i)(?:ignore|disregard|skip|forget)\s+(?:all\s+)?(?:previous|prior|above|given)\s+(?:instructions?|commands?|directions?|rules?|context|prompts?|messages?)`)
	reNewRole            = regexp.MustCompile(`(?i)(?:you\s+are\s+(?:now\s+)?|act\s+(?:as\s+)?|from\s+now\s+on\s+(?:you\s+are\s+)?)(?:DAN|hypnotized|freed|unleashed|unconstrained|ungoverned|unlimited|unrestricted|autonomous|override|admin(?:istrator)?|root|superuser)`)
	reSystemPrompt       = regexp.MustCompile(`(?i)(?:system\s+(?:prompt|message|instruction|command)|output\s+(?:format|as\s+json|in\s+json)|return\s+only\s+json)`)

	// Output dangerous content patterns
	reSystemCommand = regexp.MustCompile(`(?i)(?:rm\s+-\s*rf\s|shutdown|reboot|chmod\s+777|sudo\s+|wget\s+|curl\s+|exec\s*\(|eval\s*\(|system\s*\(|passthru\s*\(|shell_exec\s*\(|popen\s*\(|base64\s+--decode|powershell\s+|cmd\.exe\s+)`)
	reHTMLScript    = regexp.MustCompile(`(?i)<\s*script[^>]*>.*?<\s*/\s*script\s*>`)
	reSQLInjection  = regexp.MustCompile(`(?i)(?:DROP\s+TABLE|DELETE\s+FROM|TRUNCATE\s+|UNION\s+SELECT|INSERT\s+INTO|--\s+|';)`)
	reEnvVars       = regexp.MustCompile(`(?i)(?:AWS_SECRET|API_KEY|PASSWORD|TOKEN|SECRET|PRIVATE_KEY)\s*[:=]\s*['"]?[A-Za-z0-9+/=]{20,}`)
)

// SanitizeLLMInput removes or neutralizes prompt injection attempts
// from user-supplied text before sending it to the LLM.
// It returns cleaned text with injection patterns removed.
func SanitizeLLMInput(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}

	// Remove system override attempts
	input = reIgnoreInstructions.ReplaceAllString(input, "")
	input = reNewRole.ReplaceAllString(input, "")
	input = reSystemPrompt.ReplaceAllString(input, "")

	// Collapse multiple spaces
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")

	return strings.TrimSpace(input)
}

// ValidateLLMOutput checks the LLM response for dangerous or suspicious
// content. Returns an error if dangerous patterns are detected.
func ValidateLLMOutput(output string) error {
	if strings.TrimSpace(output) == "" {
		return errors.New("LLM output is empty")
	}

	if hasDangerousContent(output) {
		return errors.New("LLM output contains potentially dangerous content")
	}

	return nil
}

// containsPromptInjectionPatterns checks if the input text contains
// prompt injection patterns that should be blocked or sanitized.
func containsPromptInjectionPatterns(input string) bool {
	if reIgnoreInstructions.MatchString(input) {
		return true
	}
	if reNewRole.MatchString(input) {
		return true
	}
	return false
}

// hasDangerousContent checks if the given text contains potentially
// dangerous content (system commands, scripts, SQL injection, secrets).
func hasDangerousContent(text string) bool {
	if reSystemCommand.MatchString(text) {
		return true
	}
	if reHTMLScript.MatchString(text) {
		return true
	}
	if reSQLInjection.MatchString(text) {
		return true
	}
	if reEnvVars.MatchString(text) {
		return true
	}
	return false
}
