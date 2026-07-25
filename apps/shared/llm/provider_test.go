package llm

// Validates: REQ-FUN.API.llm-policy

import (
	"errors"
	"testing"
)

// mockProvider implements the Provider interface for testing.
type mockProvider struct {
	completeFunc       func(ctx Context, p Prompt) (Completion, error)
	streamCompleteFunc func(ctx Context, p Prompt) (<-chan Token, error)
}

func (m *mockProvider) Complete(ctx Context, p Prompt) (Completion, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, p)
	}
	return Completion{}, errors.New("Complete not implemented")
}

func (m *mockProvider) StreamComplete(ctx Context, p Prompt) (<-chan Token, error) {
	if m.streamCompleteFunc != nil {
		return m.streamCompleteFunc(ctx, p)
	}
	return nil, errors.New("StreamComplete not implemented")
}

func TestProviderInterfaceCompliance(t *testing.T) {
	// Verify mockProvider satisfies Provider interface at compile time
	var _ Provider = (*mockProvider)(nil)

	m := &mockProvider{
		completeFunc: func(ctx Context, p Prompt) (Completion, error) {
			return Completion{
				Text:         "mock response",
				TokensUsed:   10,
				FinishReason: FinishReasonStop,
			}, nil
		},
	}

	ctx := Context{}
	p := Prompt{
		Messages: []Message{{Role: RoleUser, Content: "test"}},
		Model:    "gpt-4o",
	}

	completion, err := m.Complete(ctx, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if completion.Text != "mock response" {
		t.Errorf("expected 'mock response', got '%s'", completion.Text)
	}
	if completion.TokensUsed != 10 {
		t.Errorf("expected 10 tokens, got %d", completion.TokensUsed)
	}
	if completion.FinishReason != FinishReasonStop {
		t.Errorf("expected FinishReasonStop, got %v", completion.FinishReason)
	}
}

func TestProviderStreamCompleteInterface(t *testing.T) {
	m := &mockProvider{
		streamCompleteFunc: func(ctx Context, p Prompt) (<-chan Token, error) {
			ch := make(chan Token, 3)
			ch <- Token{Text: "hello", Index: 0}
			ch <- Token{Text: " world", Index: 1}
			ch <- Token{Text: "!", Index: 2}
			close(ch)
			return ch, nil
		},
	}

	ctx := Context{}
	p := Prompt{}
	ch, err := m.StreamComplete(ctx, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var tokens []Token
	for tok := range ch {
		tokens = append(tokens, tok)
	}

	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Text != "hello" {
		t.Errorf("expected 'hello', got '%s'", tokens[0].Text)
	}
}

func TestProviderErrorPropagation(t *testing.T) {
	expectedErr := errors.New("provider unavailable")
	m := &mockProvider{
		completeFunc: func(ctx Context, p Prompt) (Completion, error) {
			return Completion{}, expectedErr
		},
	}

	_, err := m.Complete(Context{}, Prompt{})
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
