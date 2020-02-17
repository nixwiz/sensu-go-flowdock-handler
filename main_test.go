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

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	assert.Error(CheckArgs(event))
	config.FlowdockToken = "123"
	assert.Error(CheckArgs(event))
	config.AuthorName = "Sensu"
	assert.Error(CheckArgs(event))
	config.AuthorAvatar = "https://avatars1.githubusercontent.com/u/1648901?s=200&v=4"
	assert.Error(CheckArgs(event))
	config.BackendURL = "InvalidURL"
	assert.Error(CheckArgs(event))
	config.BackendURL = "http://sensu.example.com:3000"
	assert.NoError(CheckArgs(event))
}

func TestSendFlowdock(t *testing.T) {
	testcases := []struct {
		status uint32
		color  string
		state  string
	}{
		{0, "green", "OK"},
		{1, "yellow", "WARNING"},
		{2, "red", "CRITICAL"},
		{127, "orange", "UNKNOWN"},
	}

	for _, tc := range testcases {
		assert := assert.New(t)
		event := corev2.FixtureEvent("entity1", "check1")
		event.Check.Status = tc.status
		event.Check.Output = tc.state
		event.Entity.Labels = map[string]string{
			"flowdock_testlabel1": "value1",
			"flowdock_testlabel2": "value2",
		}

		var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(err)
			msg := &FlowdockMessage{}
			err = json.Unmarshal(body, msg)
			require.NoError(t, err)
			expectedTitle := tc.state + " - entity1 - check1"
			assert.Equal(expectedTitle, msg.Title)
			expectedThreadStatusColor := tc.color
			assert.Equal(expectedThreadStatusColor, msg.Thread.Status.Color)
			expectedThreadStatusValue := tc.state
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
		config.LabelPrefix = "flowdock_"
		assert.NoError(SendFlowdock(event))
	}
}

func Testmain(t *testing.T) {
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
