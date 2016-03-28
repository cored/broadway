package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/namely/broadway/broadway"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/services"
	"github.com/namely/broadway/store"
	"github.com/namely/broadway/testutils"

	"github.com/stretchr/testify/assert"
)

var testToken = "BroadwayTestToken"

func makeRequest(req *http.Request, w *httptest.ResponseRecorder) {
	mem := store.New()

	server := New(mem).Handler()
	server.ServeHTTP(w, req)
}

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

	i := map[string]interface{}{
		"playbook_id": "test",
		"id":          "test",
		"vars": map[string]string{
			"version": "ok",
		},
	}

	rbody := testutils.JsonFromMap(t, i)
	req, w := testutils.PostRequest(t, "/instances", rbody)
	makeRequest(req, w)

	assert.Equal(t, http.StatusCreated, w.Code, "Response code should be 201")
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
		rbody := testutils.JsonFromMap(t, i)
		req, w := testutils.PostRequest(t, "/instances", rbody)
		makeRequest(req, w)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected POST /instances with wrong attributes to be 400")
	}
}

func TestGetInstanceWithValidPath(t *testing.T) {
	store := store.New()
	i := broadway.Instance{PlaybookID: "foo", ID: "doesExist"}
	service := services.NewInstanceService(store)
	err := service.Create(i)
	if err != nil {
		t.Log(err.Error())
	}

	req, w := testutils.GetRequest(t, "/instance/foo/doesExist")
	makeRequest(req, w)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetInstanceWithInvalidPath(t *testing.T) {
	req, w := testutils.GetRequest(t, "/instance/foo/bar")
	makeRequest(req, w)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetInstancesWithFullPlaybook(t *testing.T) {
	testInstance1 := broadway.Instance{PlaybookID: "testPlaybookFull", ID: "testInstance1"}
	testInstance2 := broadway.Instance{PlaybookID: "testPlaybookFull", ID: "testInstance2"}
	service := services.NewInstanceService(store.New())
	err := service.Create(testInstance1)
	err = service.Create(testInstance2)
	if err != nil {
		t.Log(err.Error())
	}

	req, w := testutils.GetRequest(t, "/instances/testPlaybookFull")
	makeRequest(req, w)

	assert.Equal(t, http.StatusOK, w.Code, "Response code should be 200 OK")
}

func TestGetInstancesWithEmptyPlaybook(t *testing.T) {
	req, w := testutils.GetRequest(t, "/instances/testPlaybookFull")
	makeRequest(req, w)

	assert.Equal(t, http.StatusOK, w.Code, "Response code should be 200")
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
