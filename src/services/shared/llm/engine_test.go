package llm

// Validates: REQ-FUN.API.llm-policy

import (
	"testing"
)

func TestNewTemplateEngine(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestTemplateEngineLoadsAllTemplates(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	available := engine.Available()
	expectedCount := 5
	if len(available) != expectedCount {
		t.Errorf("expected %d templates loaded, got %d: %v", expectedCount, len(available), available)
	}
}

func TestTemplateEngineRenderOntologyGeneration(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	data := map[string]interface{}{
		"Description": "Create a person class with name and age properties.",
		"Domain":      "Human Resources",
		"Context":     "Simple ontology for employee records",
	}

	result, err := engine.Render("ontology_generation", data)
	if err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty rendered template")
	}
	if !containsStr(result, "Description") {
		t.Errorf("expected rendered template to contain Description, got:\n%s", result)
	}
}

func TestTemplateEngineRenderDocumentExtraction(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	data := map[string]interface{}{
		"Title":   "Software Architecture Document",
		"Format":  "markdown",
		"Content": "# System Architecture\n\nThe system consists of multiple microservices.",
	}

	result, err := engine.Render("document_extraction", data)
	if err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty rendered template")
	}
	if !containsStr(result, "Software Architecture Document") {
		t.Errorf("expected template to contain title, got:\n%s", result)
	}
}

func TestTemplateEngineRenderClassCompletion(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	data := map[string]interface{}{
		"OntologyContext": "Classes: Person, Organization, Address",
		"ClassName":       "Person",
		"Depth":           2,
	}

	result, err := engine.Render("class_completion", data)
	if err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty rendered template")
	}
}

func TestTemplateEngineRenderPropertySuggestion(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	data := map[string]interface{}{
		"Classes":       "Person, Organization, Address, Project",
		"SelectedClass": "Person",
	}

	result, err := engine.Render("property_suggestion", data)
	if err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty rendered template")
	}
}

func TestTemplateEngineRenderRefinement(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	data := map[string]interface{}{
		"CurrentOntology": "Classes: Person, Employee, Customer",
		"Feedback":        "Merge Employee and Customer into Person with roles",
		"Iteration":       2,
	}

	result, err := engine.Render("refinement", data)
	if err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty rendered template")
	}
}

func TestTemplateEngineUnknownTemplate(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	_, err = engine.Render("non_existent_template", nil)
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
}

func TestTemplateEngineTemplateExists(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	if !engine.TemplateExists("ontology_generation") {
		t.Error("expected ontology_generation template to exist")
	}
	if engine.TemplateExists("non_existent") {
		t.Error("expected non_existent template to not exist")
	}
}

func TestTemplateEngineValidateGoodData(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	data := map[string]interface{}{
		"Description": "Test ontology",
	}

	err = engine.ValidateTemplateData("ontology_generation", data)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestTemplateEngineValidateBadTemplate(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	err = engine.ValidateTemplateData("non_existent", nil)
	if err == nil {
		t.Fatal("expected error for non-existent template")
	}
}
