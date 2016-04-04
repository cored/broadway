package services

import (
	"fmt"

	"github.com/namely/broadway/instance"
)

// NotificationFormat how we want to display the slack notification information
const NotificationFormat = "%s-%s %s" // playbookid-id status

// NewSlackNotification builds a new slack notification
func NewSlackNotification(instance *instance.Instance) string {
	return fmt.Sprintf(
		NotificationFormat,
		instance.PlaybookID,
		instance.ID,
		instance.Status)
}
