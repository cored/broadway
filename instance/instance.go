package instance

import "encoding/json"

// Instance entity
type Instance struct {
	PlaybookID string            `json:"playbook_id" binding:"required"`
	ID         string            `json:"id"`
	Created    string            `json:"created"`
	Vars       map[string]string `json:"vars"`
	Status
}

type Status string

const (
	// StatusNew represents a newly created instance
	StatusNew Status = ""
	// StatusDeploying represents an instance that has begun deployment
	StatusDeploying = "deploying"
	// StatusDeployed represents an instance that has been deployed successfully
	StatusDeployed = "deployed"
	// StatusDeleting represents an instance that has begun deltion
	StatusDeleting = "deleting"
	// StatusError represents an instance that broke
	StatusError = "not deployed"
)

// JSON instance representation
func (i *Instance) JSON() (string, error) {
	encoded, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// Path for an instance
func (i *Instance) Path() string {
	return "/broadway/instances/" + i.PlaybookID + "/" + i.ID
}
