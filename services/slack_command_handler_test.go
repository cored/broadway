package services

import (
	"errors"
	"testing"

	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestExecuteSetvar(t *testing.T) {
	is := NewInstanceService(store.New())
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
		err := command.Execute()
		assert.Equal(t, testcase.E, err, testcase.Scenario)
		assert.Equal(t, testcase.Expected, command.Vars, testcase.Scenario)
	}
}

/*
func TestBuildSlackCommand(t *testing.T) {
	is := NewInstanceService(store.New())
	testcases := []struct {
		text         string
		expectedType SlackCommand
	}{
		{
			"dummy",
			dummyCommand,
		},
		{
			"help",
			helpCommand,
		},
		{
			"setvar",
			setvarCommand,
		},
	}
	for _, testcase := range testcases {
		c := BuildSlackCommand(testcase.text, is)
		assert.True(t, c.(type) == testcase.expectedType, "Built command was not type "+testcase.expectedType)
	}
}
*/

func TestSetvar(t *testing.T) {
	i := &instance.Instance{
		PlaybookID: "balls",
		ID:         "bowling",
	}
	is := NewInstanceService(store.New())
	is.repo.Save(i)

	c := setvarCommand{
		args: []string{"setvar", "balls", "bowling", "pins=10"},
		is:   is,
	}
	c.Execute()
	i2, _ := is.Show("balls", "bowling")
	assert.Equal(t, "10", i2.Vars["pins"], "Expected setvar to update the instance")
}
