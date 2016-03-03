package config_test

import (
	"broadway/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

var s = `
name: Service
`

func TestNewFromString(t *testing.T) {
	c, err := config.NewFromString(s)
	assert.Equal(t, nil, err)

	assert.Equal(t, `Service`, c.Name)
}
