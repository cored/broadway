package instance_test

import (
	"testing"

	. "github.com/namely/broadway/instance"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

type FakeStore struct {
	MockSetValue func(path, value string) error
	MockValue    func(path string) string
	MockValues   func(path string) map[string]string
	MockDelete   func(path string) error
}

func (fs *FakeStore) SetValue(path, value string) error {
	return fs.MockSetValue(path, value)
}

func (fs *FakeStore) Value(path string) string {
	return fs.MockValue(path)
}

func (fs *FakeStore) Values(path string) map[string]string {
	return fs.MockValues(path)
}

func (fs *FakeStore) Delete(path string) error {
	return fs.MockDelete(path)
}

func TestFindByPath(t *testing.T) {
	testcases := []struct {
		Scenario           string
		Path               Path
		Store              store.Store
		ExpectedPlaybookID string
		ExpectedError      error
	}{
		{
			Scenario: "When the instance is properly save",
			Path:     NewPath("etcdPath", "test", "id"),
			Store: &FakeStore{
				MockValue: func(path string) string {
					return `{"playbook_id":"test", "id": "id", "status": "deployed"}`
				},
			},
			ExpectedPlaybookID: "test",
			ExpectedError:      nil,
		},
		{
			Scenario: "When the instance was not properly save",
			Path:     NewPath("etcdPath", "test", "id"),
			Store: &FakeStore{
				MockValue: func(path string) string {
					return `{"playbook_id":}`
				},
			},
			ExpectedPlaybookID: "",
			ExpectedError:      ErrMalformedSaveData,
		},
		{
			Scenario: "When the instance does not exist",
			Path:     NewPath("etcdPath", "test", "id"),
			Store: &FakeStore{
				MockValue: func(path string) string {
					return ""
				},
			},
			ExpectedPlaybookID: "",
			ExpectedError:      NotFoundError("etcdPath/instances/test/id"),
		},
	}

	for _, tc := range testcases {
		returnedInstance, err := FindByPath(tc.Store, tc.Path)
		assert.Equal(t, tc.ExpectedError, err, tc.Scenario)
		if err == nil {
			assert.Equal(t, tc.ExpectedPlaybookID, returnedInstance.PlaybookID)
		}
	}
}

func TestFindByPlaybookID(t *testing.T) {
	testcases := []struct {
		Scenario          string
		Store             store.Store
		PlaybookPath      PlaybookPath
		ExpectedInstances []*Instance
		ExpectedError     error
	}{
		{
			Scenario: "When instances exist in the store",
			Store: &FakeStore{
				MockValues: func(string) map[string]string {
					return map[string]string{
						"rootPath/instances/test": `{"playbook_id": "test", "id": "id", "status": "deployed"}`,
					}
				},
			},
			PlaybookPath: NewPlaybookPath("rootPath", "test"),
			ExpectedInstances: []*Instance{
				&Instance{PlaybookID: "test", ID: "id", Status: StatusDeployed},
			},
			ExpectedError: nil,
		},
	}
	for _, tc := range testcases {
		_, err := FindByPlaybookID(tc.Store, tc.PlaybookPath)
		assert.Equal(t, tc.ExpectedError, err, tc.Scenario)
	}
}
