package manifest

import (
	"bytes"
	"text/template"
)

// Manifest represents a kubernetes manifest file
// Filename is used as identifier in the current implementation.
type Manifest struct {
	ID       string
	template *template.Template
}

// New creates a new Manifest object and parses the template
func New(id, content string) (*Manifest, error) {
	t, err := template.New(id).Parse(content)
	if err != nil {
		return nil, err
	}

	return &Manifest{ID: id, template: t}, nil
}

// Execute executes template with variables
func (m *Manifest) Execute(vars map[string]string) string {
	var b bytes.Buffer
	err := m.template.Execute(&b, vars)
	if err != nil {
		return ""
	}
	return b.String()
}
