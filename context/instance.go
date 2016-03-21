package context

import (
	"github.com/namely/broadway/domain"
	"github.com/namely/broadway/repositories"
)

type InstanceContext struct {
	repo repositories.InstanceRepository
}

func NewInstanceContext(ir repositories.InstanceRepository) *InstanceContext {
	return &InstanceContext{repo: ir}
}

func (context *InstanceContext) Create(ia domain.Instance) error {
	err := context.repo.Save(ia)
	if err != nil {
	}

	return nil
}
