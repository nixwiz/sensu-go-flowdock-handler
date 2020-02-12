package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteHandler(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Output = "OK"

	var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		msg := &FlowdockMessage{}
		err = json.Unmarshal(body, msg)
		assert.NoError(err)
		expectedTitle := "OK - entity1 - check1"
		assert.Equal(expectedTitle, msg.Title)
		expectedThreadStatusColor := "green"
		assert.Equal(expectedThreadStatusColor, msg.Thread.Status.Color)
		expectedThreadStatusValue := "OK"
		assert.Equal(expectedThreadStatusValue, msg.Thread.Status.Value)
		w.WriteHeader(http.StatusOK)
	}))

	_, err := url.ParseRequestURI(test.URL)
	require.NoError(t, err)
	config.FlowdockAPIURL = test.URL
	config.AuthorName = "Sensu"
	config.AuthorAvatar = "https://avatars1.githubusercontent.com/u/1648901?s=200&v=4"
	config.BackendURL = "http://sensu.example.com:3000"
	config.IncludeNamespace = false
	config.FlowdockToken = "123"
	assert.NoError(sendFlowdock(event))
}

func TestMain(t *testing.T) {
	assert := assert.New(t)
	file, _ := ioutil.TempFile(os.TempDir(), "sensu-go-flowdock-handler")
	defer func() {
		_ = os.Remove(file.Name())
	}()

	event := corev2.FixtureEvent("entity1", "check1")
	event.Metrics = corev2.FixtureMetrics()
	eventJSON, _ := json.Marshal(event)
	_, err := file.WriteString(string(eventJSON))
	require.NoError(t, err)
	require.NoError(t, file.Sync())
	_, err = file.Seek(0, 0)
	require.NoError(t, err)
	os.Stdin = file
	requestReceived := false

	var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"ok": true}`))
		require.NoError(t, err)
	}))

	_, err = url.ParseRequestURI(test.URL)
	require.NoError(t, err)
	oldArgs := os.Args
	os.Args = []string{"sensu-go-flowdock-handler", "--flowdockAPIURL", test.URL, "--flowdockToken", "123", "--backendURL", "http://sensu.example.com:3000"}
	defer func() { os.Args = oldArgs }()

	main()
	assert.True(requestReceived)
}
