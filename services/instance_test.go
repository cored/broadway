package services

import (
	"testing"

	"github.com/namely/broadway/broadway"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstance(t *testing.T) {
	store := store.New()
	service := NewInstanceService(store)

	i := broadway.Instance{PlaybookID: "test", ID: "222"}
	err := service.Create(i)
	assert.Nil(t, err)
	createdInstance, _ := service.Show(i.PlaybookID, i.ID)
	assert.Equal(t, "test", createdInstance.PlaybookID)
	assert.Equal(t, broadway.StatusNew, createdInstance.Status)
}

func TestShow(t *testing.T) {
	store := store.New()
	service := NewInstanceService(store)

	i := broadway.Instance{PlaybookID: "test", ID: "222"}
	err := service.Create(i)
	instance, err := service.Show(i.PlaybookID, i.ID)
	assert.Nil(t, err)
	assert.Equal(t, "test", instance.PlaybookID)
	assert.Equal(t, "222", instance.ID)
}

func TestShowMissingInstance(t *testing.T) {
	store := store.New()
	service := NewInstanceService(store)

	i := broadway.Instance{PlaybookID: "test", ID: "broken"}
	instance, err := service.Show(i.PlaybookID, i.ID)
	assert.NotNil(t, err)
	assert.Empty(t, instance.PlaybookID, "PlaybookID should be empty")
}

func TestAllWithPlaybookID(t *testing.T) {
	store := store.New()
	service := NewInstanceService(store)

	i := broadway.Instance{PlaybookID: "test", ID: "222"}
	err := service.Create(i)
	if err != nil {
		t.Log(err)
	}

	instances := service.AllWithPlaybookID(i.PlaybookID)
	assert.NotEmpty(t, instances)
}
