package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func PostRequest(t *testing.T, route string, data []byte) (*http.Request, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", route, bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
		return req, w
	}
	req.Header.Add("Content-Type", "application/json")

	return req, w
}
