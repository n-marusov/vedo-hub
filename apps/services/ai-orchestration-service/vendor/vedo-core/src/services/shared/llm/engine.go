package llm

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"strings"
	"sync"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// TemplateEngine loads and renders prompt templates.
type TemplateEngine struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewTemplateEngine creates a new template engine and loads all templates
// from the embedded filesystem.
func NewTemplateEngine() (*TemplateEngine, error) {
	engine := &TemplateEngine{
		templates: make(map[string]*template.Template),
		funcMap: template.FuncMap{
			"toUpper": strings.ToUpper,
			"toLower": strings.ToLower,
			"trim":    strings.TrimSpace,
		},
	}

	if err := engine.loadAll(); err != nil {
		return nil, fmt.Errorf("llm/templates: failed to load templates: %w", err)
	}

	return engine, nil
}

// Render renders a named template with the given data and returns the result.
func (e *TemplateEngine) Render(name string, data interface{}) (string, error) {
	e.mu.RLock()
	tmpl, ok := e.templates[name]
	e.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("llm/templates: template %q not found", name)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("llm/templates: failed to render %q: %w", name, err)
	}

	result := strings.TrimSpace(buf.String())
	log.Printf("[DEBUG] llm/templates: rendered %q (%d bytes)", name, len(result))

	return result, nil
}

// Available returns the list of loaded template names.
func (e *TemplateEngine) Available() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}

// TemplateExists returns true if a template with the given name is loaded.
func (e *TemplateEngine) TemplateExists(name string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.templates[name]
	return ok
}

// loadAll loads all .tmpl files from the embedded templates/ directory.
func (e *TemplateEngine) loadAll() error {
	entries, err := fs.ReadDir(templateFS, "templates")
	if err != nil {
		return fmt.Errorf("reading embedded templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		if err := e.load("templates/" + entry.Name()); err != nil {
			log.Printf("[WARN] llm/templates: failed to load %q: %v", entry.Name(), err)
			continue
		}
		log.Printf("[INFO] llm/templates: loaded %q", entry.Name())
	}

	return nil
}

// load loads a single template from the embedded filesystem.
func (e *TemplateEngine) load(fullPath string) error {
	content, err := templateFS.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("reading %q: %w", fullPath, err)
	}

	// Use the filename without extension as the template name
	name := strings.TrimSuffix(strings.TrimPrefix(fullPath, "templates/"), ".tmpl")

	tmpl, err := template.New(name).Funcs(e.funcMap).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing %q: %w", fullPath, err)
	}

	e.mu.Lock()
	e.templates[name] = tmpl
	e.mu.Unlock()

	return nil
}

// ValidateTemplateData checks if the given data struct can be rendered
// by the specified template.
func (e *TemplateEngine) ValidateTemplateData(name string, data interface{}) error {
	if !e.TemplateExists(name) {
		return fmt.Errorf("llm/templates: template %q not found", name)
	}

	_, err := e.Render(name, data)
	return err
}
