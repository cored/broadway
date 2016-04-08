package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	m, err := NewManifest("test", `{{ .test }}`)
	assert.Nil(t, err)

	out := m.Execute(map[string]string{"test": "hello!"})

	assert.Equal(t, "hello!", out)
}
