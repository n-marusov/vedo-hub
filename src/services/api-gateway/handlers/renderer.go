package handlers

// PromptRenderer abstracts the prompt template rendering for testability.
// The real implementation is *llm.TemplateEngine.
type PromptRenderer interface {
	Render(name string, data interface{}) (string, error)
}
