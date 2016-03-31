package services

import (
	"testing"

	"github.com/namely/broadway/broadway"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestNotify(t *testing.T) {
	testcases := []struct {
		Scenario     string
		Instance     *broadway.Instance
		SlackPayload *SlackPayload
		Expected     string
	}{
		{
			"Deployed instance",
			&broadway.Instance{PlaybookID: "mine", ID: "pr001", Status: "deployed"},
			&SlackPayload{
				Token:       "mytoken",
				TeamID:      "T0001",
				TeamDomain:  "example",
				ChannelID:   "C2147483705",
				ChannelName: "test",
				UserID:      "U2147483697",
				UserName:    "Steve",
				Command:     "/broadway",
				Text:        "deploy mine pr001",
				ResponseUrl: "https://hooks.slack.com/commands/1234/5678",
			},
			"Steve: mine-pr001 deployed",
		},
		{
			"Not deployed instance",
			&broadway.Instance{PlaybookID: "mine", ID: "pr001", Status: broadway.StatusError},
			&SlackPayload{
				Token:       "mytoken",
				TeamID:      "T0001",
				TeamDomain:  "example",
				ChannelID:   "C2147483705",
				ChannelName: "test",
				UserID:      "U2147483697",
				UserName:    "Steve",
				Command:     "/broadway",
				Text:        "deploy mine pr001",
				ResponseUrl: "https://hooks.slack.com/commands/1234/5678",
			},
			"Steve: mine-pr001 error",
		},
	}

	store := store.New()
	serviceCreator := NewInstanceService(store)

	for _, testcase := range testcases {
		err := serviceCreator.Create(testcase.Instance)
		assert.Nil(t, err)
		notifier := NewSlackDeploymentNotifier(testcase.SlackPayload, store)
		notification, err := notifier.Notify()

		assert.Nil(t, err)
		assert.Equal(t, testcase.Expected, notification, testcase.Scenario)
	}
}