package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/namely/broadway/deployment"
	"github.com/namely/broadway/env"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/store/etcdstore"
)

// SlackCommand represents a user command that came in from Slack
type SlackCommand interface {
	Execute() (string, error)
}

type unlockCommand struct {
	pID string
	ID  string
}

func (c *unlockCommand) Execute() (string, error) {
	path := instance.Path{env.EtcdPath, c.pID, c.ID}
	i, err := instance.Unlock(etcdstore.New(), path)
	if err != nil {
		return "", err
	}
	return i.String(), nil
}

type lockCommand struct {
	pID string
	ID  string
}

func (c *lockCommand) Execute() (string, error) {
	path := instance.Path{env.EtcdPath, c.pID, c.ID}
	i, err := instance.Lock(etcdstore.New(), path)
	if err != nil {
		return "", err
	}
	return i.String(), nil
}

type deployCommand struct {
	pID string
	ID  string
	is  *InstanceService
}

func (c *deployCommand) Execute() (string, error) {
	// todo: Load these from deployment package like playbooks
	ms := NewManifestService(env.ManifestsPath)
	AllManifests, err := ms.LoadManifestFolder()
	if err != nil {
		glog.Error(err)
	}

	ds := NewDeploymentService(etcdstore.New(), deployment.AllPlaybooks, AllManifests)

	i, err := c.is.Show(c.pID, c.ID)
	if err != nil {
		msg := fmt.Sprintf("Failed to deploy instance %s/%s: Instance not found", c.pID, c.ID)
		glog.Error(msg)
		return msg, err
	}

	if i.Lock {
		return i.String(), nil
	}

	go func() {
		glog.Infof("Asynchronously deploying %s/%s...", i.PlaybookID, i.ID)
		err := ds.DeployAndNotify(i)
		if err != nil {
			glog.Errorf("Slack command failed to deploy instance %s/%s:\n%s\n", i.PlaybookID, i.ID, err)
			return
		}
		glog.Infof("Slack command successfully deployed instance %s/%s", i.PlaybookID, i.ID)
		return
	}()

	return fmt.Sprintf("Started deployment of %s/%s", i.PlaybookID, i.ID), nil
}

// InvalidSetVar error presentation for invalid setvar syntax
type InvalidSetVar struct{}

func (e *InvalidSetVar) Error() string {
	return "Syntax error, example: setvar playbook1 instance10 var1=value"
}

// InvalidDeploy represents an error error for invalid deploy syntax
type InvalidDeploy struct{}

func (e *InvalidDeploy) Error() string {
	return "Syntax error, example: deploy playbook1 instance10"
}

type setvarCommand struct {
	args      []string
	playbooks map[string]*deployment.Playbook
	is        *InstanceService
}

func (c *setvarCommand) Execute() (string, error) {
	var commandMsg string
	if len(c.args) < 4 {
		return commandHints, &InvalidSetVar{}
	}
	kvs := c.args[3:] // from e.g. "setvar foo bar var1=val1 var2=val2"
	i, err := c.is.Show(c.args[1], c.args[2])
	if err != nil {
		glog.Warningf("Cannot setvars for not found instance %s/%s\n", c.args[1], c.args[2])
		return "", err
	}
	for _, kv := range kvs {
		tmp := strings.SplitN(kv, "=", 2)
		if len(tmp) != 2 {
			glog.Warningf("Setvar tried to parse badly formatted variable: %s", kv)
			return "", &InvalidSetVar{}
		}
		if c.playbookContainsVar(i.PlaybookID, tmp[0]) {
			i.Vars[tmp[0]] = tmp[1]
			commandMsg = fmt.Sprintf("Instance %s/%s updated its variables", i.PlaybookID, i.ID)
		} else {
			return fmt.Sprintf("Playbook %s does not define those variables", i.PlaybookID), &InvalidSetVar{}
		}
	}
	_, err = c.is.Update(i)
	if err != nil {
		glog.Errorf("Failed to save instance %s/%s with new vars\n", c.args[1], c.args[2])
		return "", err
	}
	return commandMsg, nil
}

func (c *setvarCommand) playbookContainsVar(playbookID, name string) bool {
	if p, ok := c.playbooks[playbookID]; ok {
		for _, playbookVar := range p.Vars {
			if playbookVar == name {
				return true
			}
		}
	}
	return false
}

// CommandHints slack commands help hints
const commandHints = `
*/bw deploy myPlaybookID myInstanceID*: Deploy an instance
*/bw info myPlaybookID myInstanceID*: Display the age and playbook variables of an instance
*/bw &lt;delete|destroy&gt; myPlaybookID myInstanceID*: Stop and remove an instance
*/bw &lt;setvar|setvars&gt; myPlaybookID myInstanceID var1=val1 ...* : Set one or more playbook variables for an instance
`

// Help slack command
type helpCommand struct{}

func (c *helpCommand) Execute() (string, error) {
	return commandHints, nil
}

// InvalidDelete definition of syntax error
type InvalidDelete struct{}

func (e *InvalidDelete) Error() string {
	return ""
}

// Delete slack command
type deleteCommand struct {
	pID string
	ID  string
	is  *InstanceService
}

func (c *deleteCommand) Execute() (string, error) {
	// todo: Load these from deployment package like playbooks
	ms := NewManifestService(env.ManifestsPath)
	AllManifests, err := ms.LoadManifestFolder()
	if err != nil {
		glog.Error(err)
	}

	ds := NewDeploymentService(etcdstore.New(), deployment.AllPlaybooks, AllManifests)

	i, err := c.is.Show(c.pID, c.ID)
	if err != nil {
		msg := fmt.Sprintf("Failed to delete instance %s/%s: Instance not found", c.pID, c.ID)
		glog.Error(msg)
		return msg, err
	}

	if i.Lock {
		return i.String(), nil
	}

	go func() {
		glog.Infof("Asynchronously deleting %s/%s...", i.PlaybookID, i.ID)

		if err := ds.DeleteAndNotify(i); err != nil {
			glog.Errorf("Slack command failed to delete instance %s/%s:\n%s\n", i.PlaybookID, i.ID, err)
			return
		}

		if err := c.is.Delete(i); err != nil {
			glog.Errorf("Slack command failed to delete instance %s/%s:\n%s\n", i.PlaybookID, i.ID, err)
			return
		}

		glog.Infof("Slack command successfully deleted instance %s/%s", i.PlaybookID, i.ID)
		return
	}()

	return fmt.Sprintf("Started deletion of %s/%s", i.PlaybookID, i.ID), nil
}

// Info slack command returns info about the instance
type infoCommand struct {
	pID string
	ID  string
	is  *InstanceService
}

func (c *infoCommand) Execute() (string, error) {
	i, err := c.is.Show(c.pID, c.ID)
	if err != nil {
		msg := fmt.Sprintf("Failed to retrieve info for %s/%s: Instance not found", c.pID, c.ID)
		glog.Error(msg)
		return msg, err
	}
	vv := varSlice{}
	for k, val := range i.Vars {
		v := varKV{
			k: k,
			v: val,
		}
		vv = append(vv, v)
	}
	sortVars(vv)
	m1 := fmt.Sprintf("Playbook: %s\n", wrapQuotes(i.PlaybookID))
	m2 := fmt.Sprintf("Instance: %s\n", wrapQuotes(i.ID))
	m3 := fmt.Sprintf("Age: %s\n", wrapQuotes(fmtAge(i.Created)))
	m4 := fmt.Sprintf("Status: %s\n", wrapQuotes(string(i.Status)))
	m5 := "Vars:\n"
	for _, vr := range vv {
		m5 += fmt.Sprintf("  - %s: %s\n", vr.k, wrapQuotes(vr.v))
	}
	msg := m1 + m2 + m3 + m4 + m5
	return msg, nil
}

func wrapQuotes(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

func fmtAge(d0 int64) string {
	t0 := time.Unix(d0, 0)
	age := time.Since(t0)
	switch {
	case age > 24*60*time.Minute:
		return fmt.Sprintf("%dd", int(age.Hours()/24))
	case age > 60*time.Minute:
		return fmt.Sprintf("%dh", int(age.Minutes()/60))
	case age > 60*time.Second:
		return fmt.Sprintf("%dm", int(age.Minutes()))
	}
	return fmt.Sprintf("%ds", int(age.Seconds()))
}

// BuildSlackCommand takes a string and some context and creates a SlackCommand
func BuildSlackCommand(payload string, is *InstanceService, playbooks map[string]*deployment.Playbook) SlackCommand {
	terms := strings.Split(payload, " ")
	switch terms[0] {
	case "setvar", "setvars": // setvar foo bar var1=val1 var2=val2
		return &setvarCommand{args: terms, is: is, playbooks: playbooks}
	case "deploy":
		if len(terms) < 3 {
			return &helpCommand{}
		}
		return &deployCommand{
			pID: terms[1],
			ID:  terms[2],
			is:  is,
		}
	case "delete", "destroy":
		if len(terms) < 3 {
			return &helpCommand{}
		}
		return &deleteCommand{pID: terms[1], ID: terms[2], is: is}
	case "info":
		if len(terms) < 3 {
			return &helpCommand{}
		}
		return &infoCommand{pID: terms[1], ID: terms[2], is: is}
	case "lock":
		if len(terms) < 3 {
			return &helpCommand{}
		}
		return &lockCommand{pID: terms[1], ID: terms[2]}
	case "unlock":
		if len(terms) < 3 {
			return &helpCommand{}
		}
		return &unlockCommand{pID: terms[1], ID: terms[2]}
	default:
		return &helpCommand{}
	}
}
