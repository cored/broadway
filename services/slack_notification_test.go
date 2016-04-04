package services

import (
	"testing"

	"github.com/namely/broadway/instance"
	"github.com/stretchr/testify/assert"
)

func TestSlackNotification(t *testing.T) {
	testcases := []struct {
		Scenario string
		Owner    SlackCommandOwner
		Instance *instance.Instance
		Expected string
	}{
		{
			"Deployed instance",
			SlackCommandOwner("Steve"),
			&instance.Instance{PlaybookID: "mine", ID: "pr001", Status: instance.StatusDeployed},
			"Steve: mine-pr001 deployed",
		},
		{
			"Not deployed instance",
			SlackCommandOwner("Bob"),
			&instance.Instance{PlaybookID: "mine", ID: "pr001", Status: instance.StatusError},
			"Bob: mine-pr001 error",
		},
	}

	for _, testcase := range testcases {
		slackNotification := NewSlackNotification(testcase.Instance, testcase.Owner)

		assert.Equal(t, testcase.Expected, slackNotification.String(), testcase.Scenario)
	}
}
