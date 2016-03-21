package domain

import "encoding/json"

type Instance struct {
	PlaybookID string            `json:"playbook_id" binding:"required"`
	ID         string            `json:"id"`
	Vars       map[string]string `json:"vars"`
}

func (instance *Instance) JSON() (string, error) {
	encoded, err := json.Marshal(instance)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func (instance *Instance) Path() string {
	return "/broadway/instances/" + instance.PlaybookID + "/" + instance.ID
}
