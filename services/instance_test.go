package services

import (
	"testing"
	"time"

	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstanceFromMissingPlaybook(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	store := store.New()
	is := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "vanishing-pb", ID: "gone"}
	_, err := is.Create(i)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "playbook vanishing-pb is missing")
}

func TestCreateInstanceWithIncorrectVars(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	store := store.New()
	is := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestCreateInstanceWithIncorrectVars", Vars: map[string]string{"metal": "plutonium"}}
	ii, err := is.Create(i)
	assert.Nil(t, ii)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not declare a var named metal")
}

func TestCreateInstanceNotification(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()

	store := store.New()
	is := NewInstanceService(store)
	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestCreateInstanceNotification"}
	_, err := is.Create(i)
	assert.Nil(t, err)

	assert.Contains(t, nt.requestBody, "created")
	assert.Contains(t, nt.requestBody, "helloplaybook")
	assert.Contains(t, nt.requestBody, "TestCreateInstanceNotification")
}

func TestCreateExistingInstanceNotChangeStatus(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	store := store.New()
	is := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestCreateInstance"}
	ii, err := is.Create(i)
	assert.Nil(t, err)
	ii.Status = instance.StatusDeployed
	updatedInstance, err := is.Update(ii)
	if err != nil {
		t.Fatal(err)
	}
	createdInstance, err := is.Create(ii)
	assert.Nil(t, err)
	assert.Equal(t, instance.StatusDeployed, createdInstance.Status)
}

func TestCreateInstanceCustomNotification(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()

	store := store.New()
	is := NewInstanceService(store)
	i := &instance.Instance{PlaybookID: "messagesplaybook", ID: "TestCreateInstanceCustomNotification"}
	_, err := is.Create(i)
	assert.Nil(t, err)

	assert.Contains(t, nt.requestBody, "custom created")
	assert.Contains(t, nt.requestBody, "messagesplaybook/TestCreateInstanceCustomNotification")
}

func TestCreateInstanceWithInvalidId(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	store := store.New()
	is := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "Test*Create_Instance"}
	ii, err := is.Create(i)
	assert.Nil(t, ii)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Test*Create_Instance")
	assert.Contains(t, err.Error(), "Test-Create-Instance")
}

func TestCreateInstance(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	store := store.New()
	is := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestCreateInstance"}
	err := is.Delete(i)
	if err != nil {
		t.Fail(err)
	}
	ii, err := is.Create(i)
	assert.Nil(t, err)
	assert.Equal(t, "helloplaybook", ii.PlaybookID)
	assert.Equal(t, instance.StatusNew, ii.Status)
	assert.Equal(t, "", ii.Vars["word"]) // Should be available since helloplaybook defines this var

	// Check the timestamp was made within the past minute:
	assert.NotEmpty(t, ii.Created)
	time0 := time.Unix(ii.Created, 0).UTC()
	time1 := time.Now().UTC()
	assert.True(t, time1.After(time0), "instance timestamp is after now")
	time2 := time0.Add(time.Minute)
	assert.True(t, time2.After(time1), "instance timestamp is more than a minute old")
}

func TestShow(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()

	store := store.New()
	is := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestShow"}
	ii, err := is.Create(i)
	assert.Nil(t, err)
	assert.Equal(t, "helloplaybook", ii.PlaybookID)
	assert.Equal(t, "TestShow", ii.ID)
}

func TestShowMissingInstance(t *testing.T) {
	store := store.New()
	is := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "broken"}
	i, err := is.Show(i.PlaybookID, i.ID)
	assert.NotNil(t, err)
	assert.Nil(t, i, "Instance should be nil")
}

func TestAllWithPlaybookID(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	is := NewInstanceService(store.New())

	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestAllWithPlaybookID"}
	_, err := is.Create(i)
	if err != nil {
		t.Fatal("TestAllWithPlaybookID: ", err)
	}

	instances, err := is.AllWithPlaybookID(i.PlaybookID)
	assert.Nil(t, err)
	assert.NotEmpty(t, instances)
	assert.Contains(t, nt.requestBody, "created")
}

func TestUpdate(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	instanceService := NewInstanceService(store.New())
	testcases := []struct {
		Scenario           string
		Instance           *instance.Instance
		ExpectedPlaybookID string
		ExpectedID         string
		ExpectedVars       map[string]string
		E                  error
	}{
		{
			"When the instance have all the needed values",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "TestUpdate"},
			"helloplaybook",
			"TestUpdate",
			map[string]string{},
			nil,
		},
	}

	for _, testcase := range testcases {
		createdInstance, err := instanceService.Create(testcase.Instance)
		if err != nil {
			t.Fatal(testcase.Scenario, err)
		}
		createdInstance.PlaybookID = testcase.ExpectedPlaybookID
		createdInstance.ID = testcase.ExpectedID
		createdInstance.Vars = testcase.ExpectedVars
		updatedInstance, err := instanceService.Update(createdInstance)

		assert.Equal(t, testcase.ExpectedPlaybookID, updatedInstance.PlaybookID)
		assert.Equal(t, testcase.E, err, testcase.Scenario)
	}
}

func TestDeleteWhenExistentInstance(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	is := NewInstanceService(store.New())

	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "new"}

	createdInstance, err := is.Create(i)
	if err != nil {
		t.Log(err)
	}
	err = is.Delete(createdInstance)
	assert.Nil(t, err, "When instance exists")
}

func TestDeleteWhenNonExistantInstance(t *testing.T) {
	is := NewInstanceService(store.New())
	i := &instance.Instance{PlaybookID: "random", ID: "bar"}

	err := is.Delete(i)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "was not found", "When non-existent instance")
}
