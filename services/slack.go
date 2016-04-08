package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/namely/broadway/deployment"
	"github.com/namely/broadway/instance"
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

// InvalidSetVar error presentation for invalid setvar syntax
type InvalidSetVar struct{}

func (ce *InvalidSetVar) Error() string {
	return "That's not the proper syntax. ex: var1=val1"
}

type setvarCommand struct {
	args      []string
	playbooks map[string]*deployment.Playbook
	is        *InstanceService
}

func (c *setvarCommand) Execute() (string, error) {
	if len(c.args) < 4 {
		return "", &InvalidSetVar{}
	}
	kvs := c.args[3:] // from e.g. "setvar foo bar var1=val1 var2=val2"
	i, err := c.is.Show(c.args[1], c.args[2])
	var commandMsg string
	if err != nil {
		glog.Warningf("Cannot setvars for not found instance %s/%s\n", c.args[1], c.args[2])
		return "", err
	}
	for _, kv := range kvs {
		tmp := strings.SplitN(kv, "=", 2)
		if len(tmp) != 2 {
			glog.Warning("Setvar tried to parse badly formatted variable: " + kv)
			return "", &InvalidSetVar{}
		}
		if c.playbookContainsVars(i) {
			if _, ok := i.Vars[tmp[0]]; ok {
				i.Vars[tmp[0]] = tmp[1]
				commandMsg = fmt.Sprintf("Instance %s %s updated it's variables",
					i.PlaybookID,
					i.ID)
			} else {
				return fmt.Sprintf("Instance %s %s does not define those variables",
					i.PlaybookID,
					i.ID), &InvalidSetVar{}
			}
		} else {
			return fmt.Sprintf("Playbook %s does not define those variables",
				i.PlaybookID), &InvalidSetVar{}
		}
	}
	_, err = c.is.Update(i)
	if err != nil {
		glog.Errorf("Failed to save instance %s/%s with new vars\n", c.args[1], c.args[2])
		return "", err
	}
	return commandMsg, nil
}

func (c *setvarCommand) playbookContainsVars(i *instance.Instance) bool {
	if _, ok := c.playbooks[i.PlaybookID]; ok {
		playbook := c.playbooks[i.PlaybookID]
		for _, playbookVar := range playbook.Vars {
			if _, ok := i.Vars[playbookVar]; ok {
				return true
			}
		}
	}
	return false
}

// CommandHints slack commands help hints
const CommandHints = `/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance`

// Help slack command
type helpCommand struct{}

func (c *helpCommand) Execute() (string, error) {
	return CommandHints, nil
}

// BuildSlackCommand takes a string and some context and creates a SlackCommand
func BuildSlackCommand(payload string, is *InstanceService, playbooks map[string]*deployment.Playbook) SlackCommand {
	terms := strings.Split(payload, " ")
	switch terms[0] {
	case "setvar": // setvar foo bar var1=val1 var2=val2
		return &setvarCommand{args: terms, is: is, playbooks: playbooks}
	default:
		return &helpCommand{}
	}
}
