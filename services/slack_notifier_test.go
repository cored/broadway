package services

import (
	"testing"

	"github.com/namely/broadway/broadway"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestNotifyWhenInstanceGetsDeployed(t *testing.T) {
	instance := &broadway.Instance{
		PlaybookID: "mine",
		ID:         "pr001",
		Status:     "deployed",
	}
	store := store.New()
	repo := broadway.NewInstanceRepo(store)
	repo.Save(instance)

	slackPayload := SlackPayload{
		Token:       "mytoken",
		TeamID:      "T0001",
		TeamDomain:  "example",
		ChannelID:   "C2147483705",
		ChannelName: "test",
		UserID:      "U2147483697",
		UserName:    "Steve",
		Command:     "/broadway deploy mine pr001",
		Text:        "94070",
		ResponseUrl: "https://hooks.slack.com/commands/1234/5678",
	}
	notifier := NewSlackDeploymentNotifier(slackPayload, store)
	notification, err := notifier.Notify()
	assert.Nil(t, err)
	assert.Equal(t, notification, "mine pr001 Steve was deployed")
}
