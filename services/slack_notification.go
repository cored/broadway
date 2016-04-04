package services

import (
	"fmt"

	"github.com/namely/broadway/instance"
)

// SlackNotification represent a notification
type SlackNotification struct {
	commandOwner SlackCommandOwner
	instance     *instance.Instance
}

func (sn SlackNotification) String() string {
	return fmt.Sprintf(
		"%s: %s-%s %s",
		sn.commandOwner,
		sn.instance.PlaybookID,
		sn.instance.ID,
		sn.instance.Status)
}

// SlackCommandOwner represents the owner for the slack command
type SlackCommandOwner string

// NewSlackNotification builds a new slack notification
func NewSlackNotification(instance *instance.Instance, owner SlackCommandOwner) SlackNotification {
	return SlackNotification{commandOwner: owner, instance: instance}
}
