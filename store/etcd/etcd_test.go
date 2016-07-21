package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	s := New()
	err := s.SetValue("/testing/a", "val")
	assert.Nil(t, err)

	val := s.Value("/testing/a")
	assert.Equal(t, "val", val)

	err = s.SetValue("/testing/a", "ok")
	assert.Nil(t, err)

	val = s.Value("/testing/a")
	assert.Equal(t, "ok", val)

	val = s.Value("/testing/empty")
	assert.Equal(t, "", val)
}

func TestValues(t *testing.T) {
	s := New()
	err := s.SetValue("/testing/vv/a", "A")
	assert.Nil(t, err)
	err = s.SetValue("/testing/vv/b", "B")
	assert.Nil(t, err)
	err = s.SetValue("/testing/vv/c", "C")
	assert.Nil(t, err)

	values := s.Values("/testing/vv")
	assert.Len(t, values, 3)
	assert.Equal(t, "A", values["a"])
	assert.Equal(t, "B", values["b"])
	assert.Equal(t, "C", values["c"])

	values = s.Values("/testing/oooo")
	assert.Len(t, values, 0)
}

func TestDelete(t *testing.T) {
	s := New()
	err := s.SetValue("/testd", "A")
	assert.Nil(t, err)

	err = s.Delete("/testd")
	assert.Nil(t, err)

	assert.Equal(t, "", s.Value("/testd"))
}
