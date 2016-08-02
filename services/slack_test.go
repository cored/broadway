package services

import (
	"testing"
	"time"

	"github.com/namely/broadway/deployment"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/store/etcdstore"
	"github.com/stretchr/testify/assert"
)

func TestDeployExecute(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	is := NewInstanceService(etcdstore.New())
	testcases := []struct {
		Scenario    string
		Arguments   string
		Instance    *instance.Instance
		Playbooks   map[string]*deployment.Playbook
		ExpectedMsg string
		E           error
	}{
		{
			"Test Deployment through slack command",
			"deploy helloplaybook chickenman",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "chickenman"},
			map[string]*deployment.Playbook{"helloplaybook": {ID: "helloplaybook"}},
			"Started deployment of helloplaybook/chickenman",
			nil,
		},
		{
			"When the instance is locked",
			"deploy helloplaybook locked",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "locked", Status: instance.StatusDeployed, Lock: true},
			map[string]*deployment.Playbook{"helloplaybook": {ID: "helloplaybook"}},
			"/broadwaytest/instances/helloplaybook/locked is currently locked",
			nil,
		},
	}

	for _, testcase := range testcases {
		_, err := is.CreateOrUpdate(testcase.Instance)
		if err != nil {
			t.Log(err)
		}
		command := BuildSlackCommand(testcase.Arguments, is, testcase.Playbooks)

		msg, err := command.Execute()
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
		assert.Equal(t, testcase.E, err, testcase.Scenario)
	}
}

func TestSetvarExecute(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()

	is := NewInstanceService(etcdstore.New())
	tPlaybooks := map[string]*deployment.Playbook{
		"helloplaybook": {
			ID:   "helloplaybook",
			Vars: []string{"word", "bird"},
		},
	}
	testcases := []struct {
		Scenario     string
		Arguments    string
		Instance     *instance.Instance
		Playbooks    map[string]*deployment.Playbook
		ExpectedVars map[string]string
		ExpectedMsg  string
		ExpectedErr  error
	}{
		{
			"When an instance sets a new variable value from the playbook",
			"setvar helloplaybook setvar1 bird=chickadee",
			&instance.Instance{
				PlaybookID: "helloplaybook",
				ID:         "setvar1",
				Vars:       map[string]string{"bird": "", "word": ""},
			},
			tPlaybooks,
			map[string]string{"bird": "chickadee", "word": ""},
			"Instance helloplaybook/setvar1 updated its variables",
			nil,
		},
		{
			"When an instance sets new variable values from the playbook",
			"setvar helloplaybook setvar2 bird=gander word=test",
			&instance.Instance{
				PlaybookID: "helloplaybook",
				ID:         "setvar2",
				Vars:       map[string]string{"bird": "", "word": ""},
			},
			tPlaybooks,
			map[string]string{"bird": "gander", "word": "test"},
			"Instance helloplaybook/setvar2 updated its variables",
			nil,
		},
		{
			"Succeeds when command is setvars",
			"setvars helloplaybook setvar2 bird=plover word=behemoth",
			&instance.Instance{
				PlaybookID: "helloplaybook",
				ID:         "setvar2",
				Vars:       map[string]string{"bird": "", "word": ""},
			},
			tPlaybooks,
			map[string]string{"bird": "plover", "word": "behemoth"},
			"Instance helloplaybook/setvar2 updated its variables",
			nil,
		},
		{
			"When the instance's playbook does not define a variable",
			"setvar helloplaybook setvar3 newvar=val1",
			&instance.Instance{
				PlaybookID: "helloplaybook",
				ID:         "setvar3",
				Vars:       map[string]string{"bird": "", "word": ""},
			},
			tPlaybooks,
			map[string]string{"bird": "", "word": ""},
			"Playbook helloplaybook does not define those variables",
			&InvalidSetVar{},
		},
		{
			"When an argument text sets 'key='",
			"setvar helloplaybook setvar4 bird=",
			&instance.Instance{
				PlaybookID: "helloplaybook",
				ID:         "setvar4",
				Vars:       map[string]string{"bird": "", "word": ""},
			},
			tPlaybooks,
			map[string]string{"bird": "", "word": ""},
			"Instance helloplaybook/setvar4 updated its variables",
			nil,
		},
		{
			"When the argument sets '=value'",
			"setvar helloplaybook setvar5 =broken",
			&instance.Instance{
				PlaybookID: "helloplaybook",
				ID:         "setvar5",
				Vars:       map[string]string{"bird": "", "word": ""},
			},
			tPlaybooks,
			map[string]string{"bird": "", "word": ""},
			"Playbook helloplaybook does not define those variables",
			&InvalidSetVar{},
		},
		{
			"When just the setvar command is sent",
			"setvar",
			&instance.Instance{
				PlaybookID: "helloplaybook",
				ID:         "setvar6",
				Vars:       map[string]string{"bird": "", "word": ""},
			},
			tPlaybooks,
			map[string]string{"bird": "", "word": ""},
			commandHints,
			&InvalidSetVar{},
		},
	}

	for _, testcase := range testcases {
		_, err := is.CreateOrUpdate(testcase.Instance)
		if err != nil {
			t.Fatal(err)
		}
		command := BuildSlackCommand(testcase.Arguments, is, testcase.Playbooks)

		msg, err := command.Execute()
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
		assert.Equal(t, testcase.ExpectedErr, err, testcase.Scenario)

		updatedInstance, err := is.Show(testcase.Instance.PlaybookID, testcase.Instance.ID)
		assert.Nil(t, err)
		assert.Equal(t, testcase.ExpectedVars, updatedInstance.Vars, testcase.Scenario)
	}
}

func TestDelete(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()

	testcases := []struct {
		Scenario    string
		Instance    *instance.Instance
		Args        string
		ExpectedMsg string
		ExpectedErr error
	}{
		{
			"Succeeds when correct delete syntax is sent",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "randomid"},
			"delete helloplaybook randomid",
			"Started deletion of helloplaybook/randomid",
			nil,
		},
		{
			"Succeeds when correct destroy syntax is sent",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "randomid"},
			"destroy helloplaybook randomid",
			"Started deletion of helloplaybook/randomid",
			nil,
		},
		{
			"When instance is locked",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "lockme", Status: instance.StatusDeployed, Lock: true},
			"destroy helloplaybook lockme",
			"/broadwaytest/instances/helloplaybook/lockme is currently locked",
			nil,
		},
		{
			"Fails when missing playbookid",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "randomid"},
			"delete randomid",
			commandHints,
			nil,
		},
	}
	is := NewInstanceService(etcdstore.New())
	for _, testcase := range testcases {
		_, err := is.CreateOrUpdate(testcase.Instance)
		if err != nil {
			t.Log(err)
		}
		command := BuildSlackCommand(
			testcase.Args,
			is,
			map[string]*deployment.Playbook{
				"helloplaybook": {ID: "randomapp"},
			},
		)

		msg, err := command.Execute()
		assert.Equal(t, testcase.ExpectedErr, err, testcase.Scenario)
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
		// Wait for Kubernetes to destroy the pod so we can recreate and destroy it in future test cases:
		time.Sleep(3 * time.Second)
	}
}

func TestHelpExecute(t *testing.T) {
	testcases := []struct {
		Scenario    string
		Args        string
		ExpectedMsg string
		ExpectedErr error
	}{
		{
			"When passing help command",
			"help",
			commandHints,
			nil,
		},
		{
			"When non existent command",
			"none",
			commandHints,
			nil,
		},
	}
	is := NewInstanceService(etcdstore.New())
	for _, testcase := range testcases {
		command := BuildSlackCommand(testcase.Args, is, nil)
		msg, err := command.Execute()
		assert.Equal(t, testcase.ExpectedErr, err, testcase.Scenario)
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
	}
}

func TestInfoExecute(t *testing.T) {
	testcases := []struct {
		Scenario    string
		Instance    *instance.Instance
		Args        string
		ExpectedMsg string
		ExpectedErr error
	}{
		{
			"info for existing instance succeeds",
			&instance.Instance{
				PlaybookID: "helloplaybook",
				ID:         "showinfo",
				Status:     instance.StatusDeployed,
				Vars:       map[string]string{"word": "phlegmatic", "bird": "albatross"},
			},
			"info helloplaybook showinfo",
			`Playbook: "helloplaybook"
Instance: "showinfo"
Age: "3s"
Status: "deployed"
Vars:
  - bird: "albatross"
  - word: "phlegmatic"
`,
			nil,
		},
		{
			"info for missing instance fails",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "randomid"},
			"info helloplaybook showmissing",
			"Failed to retrieve info for helloplaybook/showmissing: Instance not found",
			instance.NotFoundError("instances//"),
		},
		{
			"info for missing playbook fails",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "showinfo"},
			"info missingplaybook showinfo",
			"Failed to retrieve info for missingplaybook/showinfo: Instance not found",
			instance.NotFoundError("instances//"),
		},
	}
	is := NewInstanceService(etcdstore.New())
	for _, testcase := range testcases {
		_, err := is.CreateOrUpdate(testcase.Instance)
		if err != nil {
			t.Log(err)
		}
		command := BuildSlackCommand(
			testcase.Args,
			is,
			map[string]*deployment.Playbook{
				"helloplaybook": {ID: "showinfo"},
			},
		)
		// CreateOrUpdate always resets instance.Created so we can't mock it:
		time.Sleep(3 * time.Second)
		msg, err := command.Execute()
		assert.IsType(t, testcase.ExpectedErr, err, testcase.Scenario)
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
	}
}

func TestLockExecute(t *testing.T) {
	testcases := []struct {
		Scenario    string
		Instance    *instance.Instance
		Args        string
		ExpectedMsg string
		ExpectedErr error
	}{
		{
			"When passing locking command to an instance",
			&instance.Instance{

				PlaybookID: "helloplaybook",
				ID:         "lockmetest",
				Status:     instance.StatusDeployed,
				Lock:       false,
				Vars:       map[string]string{"word": "phlegmatic", "bird": "albatross"},
			},
			"lock helloplaybook lockmetest",
			"/broadwaytest/instances/helloplaybook/lockmetest is currently locked",
			nil,
		},
	}

	is := NewInstanceService(etcdstore.New())
	for _, testcase := range testcases {
		_, err := is.CreateOrUpdate(testcase.Instance)
		if err != nil {
			t.Log(err)
		}
		command := BuildSlackCommand(
			testcase.Args,
			is,
			map[string]*deployment.Playbook{
				"helloplaybook": {ID: "showinfo"},
			},
		)
		// CreateOrUpdate always resets instance.Created so we can't mock it:
		msg, err := command.Execute()
		assert.Equal(t, testcase.ExpectedErr, err, testcase.Scenario)
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
	}
}

func TestUnLockExecute(t *testing.T) {
	testcases := []struct {
		Scenario    string
		Instance    *instance.Instance
		Args        string
		ExpectedMsg string
		ExpectedErr error
	}{
		{
			"When passing unlock command to an instance",
			&instance.Instance{

				PlaybookID: "helloplaybook",
				ID:         "locktest",
				Status:     instance.StatusDeployed,
				Lock:       true,
				Vars:       map[string]string{"word": "phlegmatic", "bird": "albatross"},
			},
			"unlock helloplaybook locktest",
			"/broadwaytest/instances/helloplaybook/locktest is currently unlocked",
			nil,
		},
	}

	is := NewInstanceService(etcdstore.New())
	for _, testcase := range testcases {
		_, err := is.CreateOrUpdate(testcase.Instance)
		if err != nil {
			t.Log(err)
		}
		command := BuildSlackCommand(
			testcase.Args,
			is,
			map[string]*deployment.Playbook{
				"helloplaybook": {ID: "showinfo"},
			},
		)
		// CreateOrUpdate always resets instance.Created so we can't mock it:
		msg, err := command.Execute()
		assert.Equal(t, testcase.ExpectedErr, err, testcase.Scenario)
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
	}
}
