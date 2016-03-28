package broadway

import (
	"testing"

	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

type DummyStore struct{}

func (ds *DummyStore) Value(path string) string {
	return "malformed_json"
}

func (ds *DummyStore) SetValue(path, value string) error {
	return nil
}

func (ds *DummyStore) Values(path string) map[string]string {
	return map[string]string{"foo": "foo"}
}

func (ds *DummyStore) Delete(path string) error {
	return nil
}

func TestFindByPath(t *testing.T) {
	repo := NewInstanceRepo(store.New())
	i := Instance{PlaybookID: "test", ID: "222"}
	err := repo.Save(i)
	assert.Nil(t, err)

	instance, err := repo.FindByPath(i.Path())
	assert.Nil(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "test", instance.PlaybookID)
}

func TestFindByPathWhenTheInstanceDoesNotExist(t *testing.T) {
	repo := NewInstanceRepo(store.New())
	i := Instance{PlaybookID: "notcreated", ID: "222"}

	instance, err := repo.FindByPath(i.Path())
	assert.NotNil(t, err)
	assert.Equal(t, "Instance with path: "+i.Path()+" was not found", err.Error())
	assert.Equal(t, "", instance.PlaybookID)
	assert.Equal(t, "", instance.ID)
}

func TestFindByPathWhenMalformedData(t *testing.T) {
	repo := NewInstanceRepo(&DummyStore{})
	i := Instance{PlaybookID: "notcreated", ID: "222"}

	_, err := repo.FindByPath(i.Path())
	assert.NotNil(t, err)
	assert.Equal(t, "Saved data for this instance is malformed", err.Error())
}

func TestFindByID(t *testing.T) {
	repo := NewInstanceRepo(store.New())
	i := Instance{PlaybookID: "created", ID: "222"}

	repo.Save(i)
	instance, err := repo.FindByID(i.PlaybookID, i.ID)
	assert.Nil(t, err)
	assert.Equal(t, "created", instance.PlaybookID)
}

func TestFindByPlaybookID(t *testing.T) {
	repo := NewInstanceRepo(store.New())
	i := Instance{PlaybookID: "created", ID: "222"}
	repo.Save(i)

	instances := repo.FindByPlaybookID(i.PlaybookID)
	assert.NotNil(t, instances)
	assert.Equal(t, 1, len(instances))
}

func TestFindByPlaybookIDNoExistent(t *testing.T) {
	repo := NewInstanceRepo(store.New())

	instances := repo.FindByPlaybookID("notcreated")
	assert.NotNil(t, instances)
	assert.Equal(t, 0, len(instances))
}
