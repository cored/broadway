package services

import (
	"fmt"

	"github.com/namely/broadway/instance"
)

// NotificationFormat how we want to display the slack notification information
const NotificationFormat = "%s-%s %s"

// NewSlackNotification builds a new slack notification
func NewSlackNotification(i *instance.Instance) string {
	instanceState := i.Status
	if i.Status == instance.StatusError {
		instanceState = "failed to deploy"
	}
	return fmt.Sprintf(
		NotificationFormat,
		i.PlaybookID,
		i.ID,
		instanceState)
}
