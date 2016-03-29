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
	FindByID(playbookID, ID string) (Instance, error)
	FindByPlaybookID(playbookID string) ([]Instance, error)
}

// InstanceRepo handles persistence logic
type InstanceRepo struct {
	store store.Store
}

// NotFound instance not found error
type NotFound struct {
	path string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("Instance with path: %s was not found", e.path)
}

// MalformedSavedData malformed saved data
type MalformedSavedData struct{}

func (e MalformedSavedData) Error() string {
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
		return instance, NotFound{path}
	}
	err := json.Unmarshal([]byte(i), &instance)
	if err != nil {
		return instance, MalformedSavedData{}
	}
	return instance, nil
}

// FindByID finds an instance by playbook and instance ID
func (ir *InstanceRepo) FindByID(playbookID, ID string) (Instance, error) {
	path := "/broadway/instances/" + playbookID + "/" + ID
	return ir.FindByPath(path)
}

// FindByPlaybookID find all instances by playbook id
func (ir *InstanceRepo) FindByPlaybookID(playbookID string) ([]Instance, error) {
	data := ir.store.Values(fmt.Sprintf("/broadway/instances/%s", playbookID))
	instances := []Instance{}
	for _, value := range data {
		var instance Instance
		err := json.Unmarshal([]byte(value), &instance)
		if err != nil {
			return instances, MalformedSavedData{}
		}
		instances = append(instances, instance)
	}

	return instances, nil
}
