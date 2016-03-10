package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSavingInstance(t *testing.T) {
	i := New(&InstanceAttributes{PlaybookId: "test", Id: "222"})
	err := i.Save()
	assert.Nil(t, err)

	ni, err := Get("test", "222")
	assert.Nil(t, err)

	assert.Equal(t, "test", ni.PlaybookID())
}

func TestGettingUnsavedInstance(t *testing.T) {
	inst, err := Get("test", "none")
	assert.Nil(t, inst)
	assert.Equal(t, "Instance does not exist.", err.Error())
}

func TestDestroy(t *testing.T) {
	i := New(&InstanceAttributes{PlaybookId: "test", Id: "422"})
	assert.Nil(t, i.Save())

	assert.Nil(t, i.Destroy())

	inst, err := Get("test", "422")
	assert.Nil(t, inst)
	assert.Equal(t, "Instance does not exist.", err.Error())
}
