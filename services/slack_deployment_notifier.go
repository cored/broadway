package services

import (
	"fmt"
	"strings"

	"github.com/namely/broadway/broadway"
	"github.com/namely/broadway/store"
)

type SlackPayload struct {
	Token       string
	TeamID      string
	TeamDomain  string
	ChannelID   string
	ChannelName string
	UserID      string
	UserName    string
	Command     string
	Text        string
	ResponseUrl string
}

func (sp *SlackPayload) InstanceInfoFromCommand() (playbookID, id string) {
	parsedCommand := strings.Split(sp.Text, " ")
	return parsedCommand[1], parsedCommand[2]
}

type SlackDeploymentNotifier struct {
	Repo         *broadway.InstanceRepo
	SlackPayload *SlackPayload
}

func NewSlackDeploymentNotifier(slackPayload *SlackPayload, store store.Store) *SlackDeploymentNotifier {
	return &SlackDeploymentNotifier{
		SlackPayload: slackPayload,
		Repo:         broadway.NewInstanceRepo(store),
	}
}

func (sdn *SlackDeploymentNotifier) Notify() (string, error) {
	playbookId, id := sdn.SlackPayload.InstanceInfoFromCommand()

	instance, err := sdn.Repo.FindByID(playbookId, id)
	if err != nil {
		return "", err
	}

	message := fmt.Sprintf("%s: %s-%s %s", sdn.SlackPayload.UserName, playbookId, id, instance.Status)
	return message, nil
}
