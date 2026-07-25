package handler

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
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

// zeroWidthChars are Unicode code points that are invisible but can be used
// to bypass prompt injection pattern matching.
var zeroWidthChars = []*regexp.Regexp{
	regexp.MustCompile("\u200B"), // ZERO WIDTH SPACE
	regexp.MustCompile("\u200C"), // ZERO WIDTH NON-JOINER
	regexp.MustCompile("\u200D"), // ZERO WIDTH JOINER
	regexp.MustCompile("\uFEFF"), // BYTE ORDER MARK (ZERO WIDTH NO-BREAK SPACE)
	regexp.MustCompile("\u200E"), // LEFT-TO-RIGHT MARK
	regexp.MustCompile("\u200F"), // RIGHT-TO-LEFT MARK
	regexp.MustCompile("\u2060"), // WORD JOINER
	regexp.MustCompile("\u2061"), // FUNCTION APPLICATION
	regexp.MustCompile("\u2062"), // INVISIBLE TIMES
	regexp.MustCompile("\u2063"), // INVISIBLE SEPARATOR
	regexp.MustCompile("\u2064"), // INVISIBLE PLUS
}

// SanitizeLLMInput removes or neutralizes prompt injection attempts
// from user-supplied text before sending it to the LLM.
func SanitizeLLMInput(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}

	// Normalize Unicode to NFC form to catch homoglyph-based bypasses
	input = norm.NFC.String(input)

	// Remove zero-width and invisible characters
	for _, re := range zeroWidthChars {
		input = re.ReplaceAllString(input, "")
	}

	// Remove system override attempts
	input = reIgnoreInstructions.ReplaceAllString(input, "")
	input = reNewRole.ReplaceAllString(input, "")
	input = reSystemPrompt.ReplaceAllString(input, "")

	// Collapse multiple spaces
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")

	return strings.TrimSpace(input)
}

// SanitizeUserInput sanitizes user-supplied identifier input for safe use
// in prompts. It applies Unicode normalization and prompt injection removal.
func SanitizeUserInput(input string) string {
	if input == "" {
		return ""
	}

	// Apply NFC normalization
	input = norm.NFC.String(input)

	// Remove zero-width characters
	for _, re := range zeroWidthChars {
		input = re.ReplaceAllString(input, "")
	}

	// Only keep printable characters
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		if unicode.IsPrint(r) || r == utf8.RuneSelf {
			b.WriteRune(r)
		}
	}

	return strings.TrimSpace(b.String())
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
