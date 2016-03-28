package testutils

import (
	"encoding/json"
	"testing"
)

func JsonFromMap(t *testing.T, data map[string]interface{}) []byte {
	rbody, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
		return []byte{}
	}
	return rbody
}
