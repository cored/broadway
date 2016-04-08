package services

import (
	"testing"

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
		ExpectedVars map[string]string
		ExpectedMsg  string
		E            error
	}{
		{
			"When a playbook defines a variable for an instance",
			"setvar foo bar var1=val1",
			&instance.Instance{PlaybookID: "foo", ID: "bar", Vars: map[string]string{"var1": "val2"}},
			map[string]string{"var1": "val1"},
			"Instance foo bar updated it's variables",
			nil,
		},
		{
			"When a playbook does not defines a variable for an instance",
			"setvar foo bar newvar=val1",
			&instance.Instance{PlaybookID: "foo", ID: "bar", Vars: map[string]string{"var1": "val2"}},
			map[string]string{"var1": "val2"},
			"Instance foo bar does not define those variables",
			nil,
		},
		{
			"When an argument text just has a key",
			"setvar foobar barfoo var1=",
			&instance.Instance{PlaybookID: "foobar", ID: "barfoo"},
			nil,
			"",
			&InvalidSetVar{},
		},
		{
			"When an argument text just has a value",
			"setvar barbar foofoo =val1",
			&instance.Instance{PlaybookID: "barbar", ID: "foofoo"},
			nil,
			"",
			&InvalidSetVar{},
		},
		{
			"When just the setvar command is sent",
			"setvar",
			&instance.Instance{PlaybookID: "barbar", ID: "foofoo"},
			nil,
			"",
			&InvalidSetVar{},
		},
	}

	for _, testcase := range testcases {
		_, err := is.Create(testcase.Instance)
		command := BuildSlackCommand(testcase.Arguments, is)
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
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance`,
			nil,
		},
		{
			"When non existent command",
			"none",
			`/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance`,
			nil,
		},
	}
	is := NewInstanceService(store.New())
	for _, testcase := range testcases {
		command := BuildSlackCommand(testcase.Args, is)
		msg, err := command.Execute()
		assert.Nil(t, err)
		assert.Equal(t, testcase.Expected, msg)
	}
}
