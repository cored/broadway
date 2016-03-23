package broadway

import (
	"encoding/json"
	"fmt"

	"github.com/namely/broadway/store"
)

// InstanceRepository interface
type InstanceRepository interface {
	Save(instance Instance) error
	FindByPath(path string) (Instance, error)
	FindById(playbookId, id string) (Instance, error)
}

// InstanceRepo handles persistence logic
type InstanceRepo struct {
	store store.Store
}

// InstanceNotFoundError instance not found error
type InstanceNotFoundError struct {
	path string
}

func (e *InstanceNotFoundError) Error() string {
	return fmt.Sprintf("Instance with path: %s was not found", e.path)
}

// InstanceMalformedError instance saved with malformed data
type InstanceMalformedError struct{}

func (e *InstanceMalformedError) Error() string {
	return "Saved data for this instance is malformed"
}

// NewInstanceRepo constructor
func NewInstanceRepo(s store.Store) *InstanceRepo {
	return &InstanceRepo{store: s}
}

// Save a new instance as json
func (ir *InstanceRepo) Save(instance Instance) error {
	encoded, err := instance.JSON()
	if err != nil {
		return err
	}
	err = ir.store.SetValue(instance.Path(), encoded)
	if err != nil {
		return err
	}
	return nil
}

// FindByPath find an instance based on it's path
func (ir *InstanceRepo) FindByPath(path string) (Instance, error) {
	var instance Instance

	i := ir.store.Value(path)
	if i == "" {
		return instance, &InstanceNotFoundError{path}
	}
	err := json.Unmarshal([]byte(i), &instance)
	if err != nil {
		return instance, &InstanceMalformedError{}
	}
	return instance, nil
}

func (ir *InstanceRepo) FindById(playbookId, id string) (Instance, error) {
	path := "/broadway/instances/" + playbookId + "/" + id
	return ir.FindByPath(path)
}
