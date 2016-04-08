package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"
)

// SlackCommand represents a user command that came in from Slack
type SlackCommand interface {
	Execute() (string, error)
}

type deployCommand struct {
	args []string
}

func (c *deployCommand) Execute() error {
	return errors.New("foo")
}

type InvalidSetVar struct{}

func (ce *InvalidSetVar) Error() string {
	return "That's not the proper syntax. ex: var1=val1"
}

type setvarCommand struct {
	args []string
	is   *InstanceService
}

func (c *setvarCommand) Execute() (string, error) {
	kvs := c.args[3:len(c.args)] // from e.g. "setvar foo bar var1=val1 var2=val2"
	i, err := c.is.Show(c.args[1], c.args[2])
	if err != nil {
		glog.Warningf("Cannot setvars for not found instance %s/%s\n", c.args[1], c.args[2])
		return "", err
	}
	for _, kv := range kvs {
		kv = strings.TrimPrefix(kv, "=")
		kv = strings.TrimSuffix(kv, "=")
		tmp := strings.Split(kv, "=")
		if len(tmp) != 2 {
			glog.Warning("Setvar tried to parse badly formatted variable: " + kv)
			return "", &InvalidSetVar{}
		}
		if _, ok := i.Vars[tmp[0]]; ok {
			i.Vars[tmp[0]] = tmp[1]
		}
	}
	updatedInstance, err := c.is.Update(i)
	if err != nil {
		glog.Errorf("Failed to save instance %s/%s with new vars\n", c.args[1], c.args[2])
		return "", err
	}
	commandMsg := fmt.Sprintf("Instance %s %s updated it's variables",
		updatedInstance.PlaybookID,
		updatedInstance.ID)
	return commandMsg, nil
}

// Help slack command
type helpCommand struct{}

func (c *helpCommand) Execute() (string, error) {
	return `/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance`, nil
}

// BuildSlackCommand takes a string and some context and creates a SlackCommand
func BuildSlackCommand(payload string, is *InstanceService) SlackCommand {
	terms := strings.Split(payload, " ")
	switch terms[0] {
	case "setvar": // setvar foo bar var1=val1 var2=val2
		return &setvarCommand{args: terms, is: is}
	default:
		return &helpCommand{}
	}
	return nil
}
