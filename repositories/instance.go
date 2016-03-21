package repositories

import (
	"encoding/json"

	"github.com/namely/broadway/domain"
	"github.com/namely/broadway/store"
)

type InstanceRepository interface {
	Save(instance domain.Instance) error
	FindByPath(path string) domain.Instance
}

type InstanceRepo struct {
	store store.Store
}

func NewInstanceRepo(s store.Store) *InstanceRepo {
	return &InstanceRepo{store: s}
}

func (ir *InstanceRepo) Save(instance domain.Instance) error {
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

func (ir *InstanceRepo) FindByPath(path string) domain.Instance {
	ui, _ := json.Unmarshal([]byte(ir.store.Value(path), &domain.Instance))
	return ui
}
