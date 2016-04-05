package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/namely/broadway/env"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/playbook"
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
	env.SlackToken = testToken

	mem := store.New()

	s := New(mem)
	assert.Equal(t, testToken, s.slackToken, "Expected server.slackToken to match existing ENV value")

	env.SlackToken = ""
	s = New(mem)
	assert.Equal(t, "", s.slackToken, "Expected server.slackToken to be empty string")

}

func TestInstanceCreateWithValidAttributes(t *testing.T) {

	i := map[string]interface{}{
		"playbook_id": "test",
		"id":          "test",
		"vars": map[string]string{
			"version": "ok",
		},
	}

	rbody := testutils.JSONFromMap(t, i)
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
		rbody := testutils.JSONFromMap(t, i)
		req, w := testutils.PostRequest(t, "/instances", rbody)
		makeRequest(req, w)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected POST /instances with wrong attributes to be 400")
	}
}

func TestGetInstanceWithValidPath(t *testing.T) {
	store := store.New()
	i := &instance.Instance{PlaybookID: "foo", ID: "doesExist"}
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
	testInstance1 := &instance.Instance{PlaybookID: "testPlaybookFull", ID: "testInstance1"}
	testInstance2 := &instance.Instance{PlaybookID: "testPlaybookFull", ID: "testInstance2"}
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
			"/status/goodPlaybook/badInstance",
			404,
			"Not Found",
		},
	}

	for _, i := range invalidRequests {
		req, w := testutils.GetRequest(t, i.path)
		makeRequest(req, w)

		assert.Equal(t, i.errCode, w.Code)

		var errorResponse map[string]string

		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.Nil(t, err)
		assert.Contains(t, errorResponse["error"], i.errMsg)
	}

}
func TestGetStatusWithGoodPath(t *testing.T) {
	testInstance1 := &instance.Instance{
		PlaybookID: "goodPlaybook",
		ID:         "goodInstance",
		Status:     instance.StatusDeployed}
	service := services.NewInstanceService(store.New())
	err := service.Create(testInstance1)
	if err != nil {
		t.Error(err)
		return
	}
	req, w := testutils.GetRequest(t, "/status/goodPlaybook/goodInstance")
	makeRequest(req, w)

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
	env.SlackToken = testToken
	w, server := helperSetupServer()
	formBytes := bytes.NewBufferString("not a form")
	req, _ := http.NewRequest("POST", "/command", formBytes)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected POST /command with bad body to be 401")
}
func TestPostCommandWrongToken(t *testing.T) {
	env.SlackToken = testToken
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
	env.SlackToken = testToken
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
func TestPostCommandDeployBad(t *testing.T) {
	env.SlackToken = testToken
	w, server := helperSetupServer()
	req, _ := http.NewRequest("POST", "/command", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form := url.Values{}
	form.Set("token", testToken)
	form.Set("command", "/broadway")
	form.Set("text", "deploy foo")
	req.PostForm = form

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected /broadway deploy foo to be 200")
	assert.Contains(t, w.Body.String(), "/broadway deploy myPlaybookID myInstanceID", "Expected help message to contain /broadway")
}

func TestDeployMissing(t *testing.T) {
	mem := store.New()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/deploy/missingPlaybook/missingInstance", nil)
	assert.Nil(t, err)

	server := New(mem).Handler()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse map[string]string
	log.Println(w.Body.String())
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.Nil(t, err)
	assert.Contains(t, errorResponse["error"], "Not Found")
}

func TestDeployGood(t *testing.T) {
	task := playbook.Task{
		Name: "First step",
		Manifests: []string{
			"test",
		},
	}
	// Ensure playbook is in memory
	p := &playbook.Playbook{
		ID:    "test",
		Name:  "Test deployment",
		Meta:  playbook.Meta{},
		Vars:  []string{"test"},
		Tasks: []playbook.Task{task},
	}

	// Ensure manifest "test.yml" present in manifests folder
	// Setup server
	mem := store.New()
	server := New(mem)
	server.playbooks = map[string]*playbook.Playbook{p.ID: p}
	// engine := server.Handler()

	// Ensure instance present in etcd
	// Call endpoint
	// Assert successful deploy
	// Teardown kubernetes topography
}
