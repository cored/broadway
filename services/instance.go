package services

import (
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/store"
)

// InstanceService definition
type InstanceService struct {
	repo instance.Repository
}

// NewInstanceService creates a new instance service
func NewInstanceService(s store.Store) *InstanceService {
	return &InstanceService{repo: instance.NewRepo(s)}
}

// Create a new instance
func (is *InstanceService) Create(i *instance.Instance) (*instance.Instance, error) {
	err := is.repo.Save(i)
	if err != nil {
		return nil, err
	}
	createdInstance, err := is.repo.FindByPath(i.Path())
	if err != nil {
		return nil, err
	}
	return createdInstance, nil
}

// Update an instance
func (is *InstanceService) Update(instance *instance.Instance) (*instance.Instance, error) {
	return is.Create(instance)
}

// Show takes playbookID and instanceID and returns the matching Instance, if
// any
func (is *InstanceService) Show(playbookID, ID string) (*instance.Instance, error) {
	instance, err := is.repo.FindByID(playbookID, ID)
	if err != nil {
		return instance, err
	}
	return instance, nil
}

// AllWithPlaybookID returns all the instances for an specified playbook id
func (is *InstanceService) AllWithPlaybookID(playbookID string) ([]*instance.Instance, error) {
	return is.repo.FindByPlaybookID(playbookID)
}
