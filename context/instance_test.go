package context

import (
	"testing"

	"github.com/namely/broadway/domain"
	"github.com/namely/broadway/repositories"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	store := store.New()
	repo := repositories.NewInstanceRepo(store)
	context := NewInstanceContext(repo)

	ia := domain.Instance{PlaybookID: "test", ID: "222"}
	err := context.Create(ia)

	assert.Nil(t, err)
	assert.Equal(t, "test", repo.FindByPath(instance.Path()).PlaybookID)
}
