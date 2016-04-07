package services

import (
	"errors"
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
		Scenario  string
		Arguments []string
		Expected  map[string]string
		E         error
	}{
		{
			"When a valid set of arguments been passed",
			[]string{"setvar", "foo", "bar", "var1=val1"},
			map[string]string{"var1": "val1"},
			nil,
		},
		{
			"When an argument text just has a key",
			[]string{"setvar", "foo", "bar", "var1="},
			map[string]string{},
			errors.New("This is not the proper syntax. ex: var1=val1"),
		},
		{
			"When an argument text just has a value",
			[]string{"setvar", "foo", "bar", "=val1"},
			map[string]string{},
			errors.New("This is not the proper syntax. ex: var1=val1"),
		},
	}

	for _, testcase := range testcases {
		command := &setvarCommand{
			args: testcase.Arguments,
			is:   is,
		}
		msg, err := command.Execute()
		assert.Empty(t, msg)
		assert.Equal(t, testcase.E, err, testcase.Scenario)
		assert.Equal(t, testcase.Expected, command.Vars, testcase.Scenario)
	}
}

func TestHelpExecute(t *testing.T) {
	testcases := []struct {
		Scenario string
		Expected string
		E        error
	}{
		{
			"When passing help command",
			`/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance`,
			nil,
		},
	}

	for _, testcase := range testcases {
		command := &helpCommand{}
		msg, err := command.Execute()
		assert.Nil(t, err)
		assert.Equal(t, testcase.Expected, msg)
	}
}
