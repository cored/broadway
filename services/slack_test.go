package services

import (
	"testing"

	"github.com/namely/broadway/deployment"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestSetvarExecute(t *testing.T) {
	is := NewInstanceService(store.New())
	i := &instance.Instance{
		PlaybookID: "foo",
		ID:         "bar",
	}
	is.repo.Save(i)
	testcases := []struct {
		Scenario     string
		Arguments    string
		Instance     *instance.Instance
		Playbooks    map[string]*deployment.Playbook
		ExpectedVars map[string]string
		ExpectedMsg  string
		E            error
	}{
		{
			"When an instance and playbook defines a variable",
			"setvar foo bar var1=val1",
			&instance.Instance{PlaybookID: "foo", ID: "bar", Vars: map[string]string{"var1": "val2"}},
			map[string]*deployment.Playbook{"foo": &deployment.Playbook{ID: "foo", Vars: []string{"var1"}}},
			map[string]string{"var1": "val1"},
			"Instance foo bar updated it's variables",
			nil,
		},
		{
			"When an instance and playbook defines the same variables",
			"setvar foo bar var1=val1 var2=val2",
			&instance.Instance{PlaybookID: "foo", ID: "bar", Vars: map[string]string{"var1": "val2", "var2": "val1"}},
			map[string]*deployment.Playbook{"foo": &deployment.Playbook{ID: "foo", Vars: []string{"var1", "var2"}}},
			map[string]string{"var1": "val1", "var2": "val2"},
			"Instance foo bar updated it's variables",
			nil,
		},
		{
			"When the instance's playbook does not defines a variable",
			"setvar foo bar newvar=val1",
			&instance.Instance{PlaybookID: "foo", ID: "bar", Vars: map[string]string{"newvar": "val2"}},
			map[string]*deployment.Playbook{"foo": &deployment.Playbook{ID: "foo", Vars: []string{""}}},
			map[string]string{"newvar": "val2"},
			"Playbook foo does not define those variables",
			&InvalidSetVar{},
		},
		{
			"When an argument text just has a key",
			"setvar foobar barfoo var1=",
			&instance.Instance{PlaybookID: "foobar", ID: "barfoo"},
			map[string]*deployment.Playbook{"foo": &deployment.Playbook{ID: "foobar", Vars: []string{"var1"}}},
			nil,
			"Playbook foobar does not define those variables",
			&InvalidSetVar{},
		},
		{
			"When the argument just have a value",
			"setvar barbar foofoo =val1",
			&instance.Instance{PlaybookID: "barbar", ID: "foofoo"},
			map[string]*deployment.Playbook{"foo": &deployment.Playbook{ID: "foo", Vars: []string{"var1"}}},
			nil,
			"Playbook barbar does not define those variables",
			&InvalidSetVar{},
		},
		{
			"When just the setvar command is sent",
			"setvar",
			&instance.Instance{PlaybookID: "barbar", ID: "foofoo"},
			map[string]*deployment.Playbook{"foo": &deployment.Playbook{ID: "foo", Vars: []string{""}}},
			nil,
			"",
			&InvalidSetVar{},
		},
	}

	for _, testcase := range testcases {
		_, err := is.Create(testcase.Instance)
		command := BuildSlackCommand(testcase.Arguments, is, testcase.Playbooks)
		if err != nil {
			t.Log(err)
		}

		msg, err := command.Execute()
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
		assert.Equal(t, testcase.E, err, testcase.Scenario)

		updatedInstance, _ := is.Show(testcase.Instance.PlaybookID, testcase.Instance.ID)
		assert.Equal(t, testcase.ExpectedVars, updatedInstance.Vars, testcase.Scenario)
	}
}

func TestHelpExecute(t *testing.T) {
	testcases := []struct {
		Scenario string
		Args     string
		Expected string
		E        error
	}{
		{
			"When passing help command",
			"help",
			`/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance
/broadway setvar myPlaybookID myInstanceID var1=val1 githash=8ad33dad env=prod`,
			nil,
		},
		{
			"When non existent command",
			"none",
			`/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance
/broadway setvar myPlaybookID myInstanceID var1=val1 githash=8ad33dad env=prod`,
			nil,
		},
	}
	is := NewInstanceService(store.New())
	for _, testcase := range testcases {
		command := BuildSlackCommand(testcase.Args, is, nil)
		msg, err := command.Execute()
		assert.Nil(t, err)
		assert.Equal(t, testcase.Expected, msg)
	}
}
