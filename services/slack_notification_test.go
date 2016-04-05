package services

import (
	"testing"

	"github.com/namely/broadway/instance"
	"github.com/stretchr/testify/assert"
)

func TestSlackNotification(t *testing.T) {
	testcases := []struct {
		Scenario string
		Instance *instance.Instance
		Expected string
	}{
		{
			"Deployed instance",
			&instance.Instance{PlaybookID: "mine", ID: "pr001", Status: instance.StatusDeployed},
			"mine-pr001 deployed",
		},
		{
			"Not deployed instance",
			&instance.Instance{PlaybookID: "mine", ID: "pr001", Status: instance.StatusError},
			"mine-pr001 failed to deploy",
		},
	}

	for _, testcase := range testcases {
		slackNotification := NewSlackNotification(testcase.Instance)

		assert.Equal(t, testcase.Expected, slackNotification, testcase.Scenario)
	}
}
