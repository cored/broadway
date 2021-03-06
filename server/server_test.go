package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/services"
	"github.com/namely/broadway/store"

	"github.com/stretchr/testify/assert"
)

var testToken = "BroadwayTestToken"

func TestServerNew(t *testing.T) {
	err := os.Setenv(slackTokenENV, testToken)
	if err != nil {
		t.Fatal(err)
	}
	actualToken, exists := os.LookupEnv(slackTokenENV)
	assert.True(t, exists, "Expected ENV to exist")
	assert.Equal(t, testToken, actualToken, "Unexpected ENV value")

	mem := store.New()

	s := New(mem)
	assert.Equal(t, testToken, s.slackToken, "Expected server.slackToken to match existing ENV value")

	err = os.Unsetenv(slackTokenENV)
	if err != nil {
		t.Fatal(err)
	}
	actualToken, exists = os.LookupEnv(slackTokenENV)
	assert.False(t, exists, "Expected ENV to not exist")
	assert.Equal(t, "", actualToken, "Unexpected ENV value")
	s = New(mem)
	assert.Equal(t, "", s.slackToken, "Expected server.slackToken to be empty string for missing ENV value")

}

func TestInstanceCreateWithValidAttributes(t *testing.T) {
	w := httptest.NewRecorder()

	i := map[string]interface{}{
		"playbook_id": "test",
		"id":          "test",
		"vars": map[string]string{
			"version": "ok",
		},
	}

	rbody, err := json.Marshal(i)
	if err != nil {
		t.Error(err)
		return
	}

	req, err := http.NewRequest("POST", "/instances", bytes.NewBuffer(rbody))
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	mem := store.New()

	server := New(mem).Handler()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Response code should be 201")

	var response instance.Attributes
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, response.PlaybookID, "test")

	service := services.NewInstanceService(mem)
	ii, err := service.Show("test", "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", ii.ID, "New instance was created")

}

func TestCreateInstanceWithInvalidAttributes(t *testing.T) {

	invalidRequests := map[string]map[string]interface{}{
		"playbook_id": {
			"id": "test",
			"vars": map[string]string{
				"version": "ok",
			},
		},
	}

	for _, i := range invalidRequests {
		w := httptest.NewRecorder()
		rbody, err := json.Marshal(i)
		if err != nil {
			t.Error(err)
			return
		}

		req, err := http.NewRequest("POST", "/instances", bytes.NewBuffer(rbody))
		if err != nil {
			t.Error(err)
			return
		}

		req.Header.Add("Content-Type", "application/json")

		mem := store.New()

		server := New(mem).Handler()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected POST /instances with wrong attributes to be 400")

		var errorResponse map[string]string

		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Contains(t, errorResponse["error"], "Missing")
	}

}

func TestGetInstanceWithValidPath(t *testing.T) {
	w := httptest.NewRecorder()
	mem := store.New()

	i := instance.New(mem, &instance.Attributes{
		PlaybookID: "foo",
		ID:         "doesExist",
	})
	err := i.Save()
	if err != nil {
		t.Error(err)
		return
	}

	req, err := http.NewRequest("GET", "/instance/foo/doesExist", nil)
	if err != nil {
		t.Error(err)
		return
	}

	server := New(mem).Handler()
	server.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)

	var iResponse map[string]string

	err = json.Unmarshal(w.Body.Bytes(), &iResponse)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Contains(t, iResponse["id"], "doesExist")
}

func TestGetInstanceWithInvalidPath(t *testing.T) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/instance/foo/bar", nil)
	if err != nil {
		t.Error(err)
		return
	}

	mem := store.New()

	server := New(mem).Handler()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse map[string]string

	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Contains(t, errorResponse["error"], "Not Found")
}

func TestGetInstancesWithFullPlaybook(t *testing.T) {
	w := httptest.NewRecorder()
	mem := store.New()

	testInstance1 := instance.New(mem, &instance.Attributes{
		PlaybookID: "testPlaybookFull",
		ID:         "testInstance1",
	})
	err := testInstance1.Save()
	if err != nil {
		t.Error(err)
		return
	}
	testInstance2 := instance.New(mem, &instance.Attributes{
		PlaybookID: "testPlaybookFull",
		ID:         "testInstance2",
	})
	err = testInstance2.Save()
	if err != nil {
		t.Error(err)
		return
	}
	req, err := http.NewRequest("GET", "/instances/testPlaybookFull", nil)
	if err != nil {
		t.Error(err)
		return
	}

	server := New(mem).Handler()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response code should be 200 OK")

	log.Println(w.Body.String())
	var okResponse []instance.Attributes

	err = json.Unmarshal(w.Body.Bytes(), &okResponse)
	if err != nil {
		t.Error(err)
		return
	}
	if len(okResponse) != 2 {
		t.Errorf("Expected 2 instances matching playbook testPlaybookFull, actual %v\n", len(okResponse))
	}
}

func TestGetInstancesWithEmptyPlaybook(t *testing.T) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/instances/testPlaybookEmpty", nil)
	if err != nil {
		t.Error(err)
		return
	}

	mem := store.New()

	server := New(mem).Handler()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code, "Response code should be 204 No Content")

	var okResponse []instance.Attributes

	err = json.Unmarshal(w.Body.Bytes(), &okResponse)
	if err != nil {
		t.Error(err)
		return
	}
	if len(okResponse) != 0 {
		t.Errorf("Expected 0 instances matching playbook testPlaybookEmpty, actual %v\n", len(okResponse))
	}
}

func TestGetStatusFailures(t *testing.T) {
	invalidRequests := []struct {
		method  string
		path    string
		errCode int
		errMsg  string
	}{
		{
			"GET",
			"/status",
			400,
			"Use GET /status/yourPlaybookId/yourInstanceId",
		},
		{
			"GET",
			"/status/goodPlaybook",
			400,
			"Use GET /status/yourPlaybookId/yourInstanceId",
		},
		/* TODO: Store and look up playbooks
		{
			"GET",
			"/status/badPlaybook/goodInstance",
			404,
			"Playbook badPlaybook not found",
		},
		*/
		{
			"GET",
			"/status/goodPlaybook/badInstance",
			404,
			"Not Found",
		},
	}

	mem := store.New()
	server := New(mem).Handler()

	for _, i := range invalidRequests {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", i.path, nil)
		assert.Nil(t, err)

		server.ServeHTTP(w, req)

		assert.Equal(t, i.errCode, w.Code)

		var errorResponse map[string]string

		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.Nil(t, err)
		assert.Contains(t, errorResponse["error"], i.errMsg)
	}

}
func TestGetStatusWithGoodPath(t *testing.T) {
	mem := store.New()
	testInstance1 := instance.New(mem, &instance.Attributes{
		PlaybookID: "goodPlaybook",
		ID:         "goodInstance",
		Status:     instance.StatusDeployed,
	})
	err := testInstance1.Save()
	if err != nil {
		t.Error(err)
		return
	}
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/status/goodPlaybook/goodInstance", nil)
	assert.Nil(t, err)

	server := New(mem).Handler()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var statusResponse map[string]string

	err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
	assert.Nil(t, err)
	assert.Contains(t, statusResponse["status"], "deployed")
}

func helperSetupServer() (*httptest.ResponseRecorder, http.Handler) {
	w := httptest.NewRecorder()
	mem := store.New()
	server := New(mem).Handler()
	return w, server
}

func TestGetCommand400(t *testing.T) {
	w, server := helperSetupServer()
	req, err := http.NewRequest("GET", "/command", nil)
	if err != nil {
		t.Fatal(err)
	}

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected GET /command to be 400")
}
func TestGetCommand200(t *testing.T) {
	w, server := helperSetupServer()
	req, _ := http.NewRequest("GET", "/command?ssl_check=1", nil)

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected GET /command?ssl_check=1 to be 200")
}
func TestPostCommandMissingToken(t *testing.T) {
	if err := os.Setenv(slackTokenENV, testToken); err != nil {
		t.Fatal(err)
	}
	w, server := helperSetupServer()
	formBytes := bytes.NewBufferString("not a form")
	req, _ := http.NewRequest("POST", "/command", formBytes)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected POST /command with bad body to be 401")
}
func TestPostCommandWrongToken(t *testing.T) {
	if err := os.Setenv(slackTokenENV, testToken); err != nil {
		t.Fatal(err)
	}
	w, server := helperSetupServer()
	req, _ := http.NewRequest("POST", "/command", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form := url.Values{}
	form.Set("token", "wrongtoken")
	req.PostForm = form

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected POST /command with wrong token to be 401")
}
func TestPostCommandHelp(t *testing.T) {
	if err := os.Setenv(slackTokenENV, testToken); err != nil {
		t.Fatal(err)
	}
	w, server := helperSetupServer()
	req, _ := http.NewRequest("POST", "/command", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form := url.Values{}
	form.Set("token", testToken)
	form.Set("command", "/broadway")
	form.Set("text", "help")
	req.PostForm = form

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected /broadway help to be 200")
	assert.Contains(t, w.Body.String(), "/broadway", "Expected help message to contain /broadway")
}
func TestPostCommand(t *testing.T) {
}
