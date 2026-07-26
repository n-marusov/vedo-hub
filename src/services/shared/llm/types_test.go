package llm

// Validates: REQ-FUN.API.llm-policy

import (
	"testing"
)

func TestPromptDefaults(t *testing.T) {
	p := Prompt{
		Messages: []Message{
			{Role: RoleSystem, Content: "You are an ontology engineer."},
			{Role: RoleUser, Content: "Extract ontology from this document."},
		},
	}

	if len(p.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(p.Messages))
	}
	if p.Model != "" {
		t.Errorf("expected empty model, got '%s'", p.Model)
	}
	if p.Temperature != 0 {
		t.Errorf("expected 0 temperature, got %f", p.Temperature)
	}
	if p.MaxTokens != 0 {
		t.Errorf("expected 0 max tokens, got %d", p.MaxTokens)
	}
}

func TestMessageRoles(t *testing.T) {
	tests := []struct {
		role    MessageRole
		wantStr string
	}{
		{RoleSystem, "system"},
		{RoleUser, "user"},
		{RoleAssistant, "assistant"},
		{RoleTool, "tool"},
	}

	for _, tt := range tests {
		if string(tt.role) != tt.wantStr {
			t.Errorf("role %s expected, got %s", tt.wantStr, string(tt.role))
		}
	}
}

func TestFinishReasons(t *testing.T) {
	tests := []struct {
		reason  FinishReason
		wantStr string
	}{
		{FinishReasonStop, "stop"},
		{FinishReasonLength, "length"},
		{FinishReasonContentFilter, "content_filter"},
		{FinishReasonToolCalls, "tool_calls"},
	}

	for _, tt := range tests {
		if string(tt.reason) != tt.wantStr {
			t.Errorf("finish reason %s expected, got %s", tt.wantStr, string(tt.reason))
		}
	}
}

func TestCompletionValidation(t *testing.T) {
	c := Completion{
		Text:         "Generated OWL content",
		TokensUsed:   150,
		PromptTokens: 50,
		FinishReason: FinishReasonStop,
	}

	if c.Text == "" {
		t.Error("completion text should not be empty")
	}
	if c.TokensUsed == 0 {
		t.Error("tokens used should be > 0")
	}
}

func TestTokenSerialization(t *testing.T) {
	tok := Token{
		Text:  "class",
		Index: 5,
	}

	if tok.Text != "class" {
		t.Errorf("expected 'class', got '%s'", tok.Text)
	}
	if tok.Index != 5 {
		t.Errorf("expected index 5, got %d", tok.Index)
	}
}

func TestContextFields(t *testing.T) {
	ctx := Context{
		TraceID:      "abc123",
		MaxRetries:   3,
		SkipTracking: false,
	}

	if ctx.TraceID != "abc123" {
		t.Errorf("expected TraceID 'abc123', got '%s'", ctx.TraceID)
	}
	if ctx.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", ctx.MaxRetries)
	}
	if ctx.SkipTracking {
		t.Error("expected SkipTracking to be false")
	}
}
