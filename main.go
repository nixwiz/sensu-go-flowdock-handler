package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// FlowdockMessageAuthor is a subtype in FlowdockMessage
type FlowdockMessageAuthor struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// FlowdockMessageThreadStatus is a subtype in FlowdockMessageThread
type FlowdockMessageThreadStatus struct {
	Color string `json:"color"`
	Value string `json:"value"`
}

// FlowdockMessageThreadFields is a subtype in FlowdockMessageThread
type FlowdockMessageThreadFields struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// FlowdockMessageThread is is a subtype in FlowdockMessage
type FlowdockMessageThread struct {
	Title       string                        `json:"title"`
	Fields      []FlowdockMessageThreadFields `json:"fields"`
	Body        string                        `json:"body"`
	ExternalURL string                        `json:"external_url"`
	Status      FlowdockMessageThreadStatus   `json:"status"`
}

// FlowdockMessage is the JSON type for a Flowdock message
type FlowdockMessage struct {
	Flowtoken        string                `json:"flow_token"`
	Event            string                `json:"event"`
	Author           FlowdockMessageAuthor `json:"author"`
	Title            string                `json:"title"`
	ExternalThreadID string                `json:"external_thread_id"`
	Thread           FlowdockMessageThread `json:"thread"`
}

const (
	flowdockToken    = "flowdockToken"
	authorName       = "authorName"
	authorAvatar     = "authorAvatar"
	backendURL       = "backendURL"
	includeNamespace = "includeNamespace"
	labelPrefix      = "labelPrefix"
	flowdockAPIURL   = "flowdockAPIURL"
)

// HandlerConfig is needed for Sensu Go Handlers
type HandlerConfig struct {
	sensu.PluginConfig
	FlowdockToken    string
	AuthorName       string
	AuthorAvatar     string
	BackendURL       string
	LabelPrefix      string
	IncludeNamespace bool
	FlowdockAPIURL   string
}

var (
	threadBody           string
	msgTitle             string
	msgThreadTitle       string
	msgThreadExternalURL string
	msgThreadStatusColor string
	msgThreadStatusValue string

	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-go-flowdock-handler",
			Short:    "The Sensu Go Flowdock handler for sending notifications",
			Keyspace: "sensu.io/plugins/flowdock/config",
		},
	}
	flowdockConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      flowdockToken,
			Env:       "SENSU_FLOWDOCK_TOKEN",
			Argument:  flowdockToken,
			Shorthand: "t",
			Default:   "",
			Usage:     "The Flowdock application token",
			Value:     &config.FlowdockToken,
		},
		{
			Path:      authorName,
			Argument:  authorName,
			Shorthand: "n",
			Default:   "Sensu",
			Usage:     "Name for the author of the thread",
			Value:     &config.AuthorName,
		},
		{
			Path:      authorAvatar,
			Argument:  authorAvatar,
			Shorthand: "a",
			Default:   "https://avatars1.githubusercontent.com/u/1648901?s=200&v=4",
			Usage:     "Avatar URL",
			Value:     &config.AuthorAvatar,
		},
		{
			Path:      backendURL,
			Env:       "SENSU_FLOWDOCK_BACKENDURL",
			Argument:  backendURL,
			Shorthand: "b",
			Default:   "",
			Usage:     "The URL for the backend, used to create links to events",
			Value:     &config.BackendURL,
		},
		{
			Path:      labelPrefix,
			Argument:  labelPrefix,
			Shorthand: "l",
			Default:   "",
			Usage:     "Label prefix for entity fields to be included in thread",
			Value:     &config.LabelPrefix,
		},
		{
			Path:      includeNamespace,
			Argument:  includeNamespace,
			Shorthand: "i",
			Default:   false,
			Usage:     "Include the namespace with the entity name in title and thread ID",
			Value:     &config.IncludeNamespace,
		},
		{
			Path:      flowdockAPIURL,
			Argument:  flowdockAPIURL,
			Shorthand: "u",
			Default:   "https://api.flowdock.com/messages",
			Usage:     "The Flowdock API URL",
			Value:     &config.FlowdockAPIURL,
		},
	}
)

func main() {

	goHandler := sensu.NewGoHandler(&config.PluginConfig, flowdockConfigOptions, checkArgs, sendFlowdock)
	goHandler.Execute()

}

func checkArgs(_ *corev2.Event) error {

	if len(config.FlowdockToken) == 0 {
		return errors.New("missing Flowdock token")
	}
	if len(config.AuthorName) == 0 {
		return errors.New("missing author name supecification")
	}
	if len(config.AuthorAvatar) == 0 {
		return errors.New("missing author avatar URL specification")
	}
	if len(config.BackendURL) == 0 {
		return errors.New("missing backend URL specification")
	}
	if !govalidator.IsURL(config.BackendURL) {
		return errors.New("invlaid backend URL specification")
	}
	config.BackendURL = strings.TrimSuffix(config.BackendURL, "/")

	return nil
}

func sendFlowdock(event *corev2.Event) error {

	var (
		msgThreadStatusColor string
		msgThreadStatusValue string
		msgNamespace         string
	)

	switch eventStatus := event.Check.Status; eventStatus {
	case 0:
		msgThreadStatusColor = "green"
		msgThreadStatusValue = "OK"
	case 1:
		msgThreadStatusColor = "yellow"
		msgThreadStatusValue = "WARNING"
	case 2:
		msgThreadStatusColor = "red"
		msgThreadStatusValue = "CRITICAL"
	default:
		msgThreadStatusColor = "orange"
		msgThreadStatusValue = "UNKNOWN"
	}

	if config.IncludeNamespace {
		msgNamespace = event.Entity.Namespace + "/"
	} else {
		msgNamespace = ""
	}

	msgThreadExternalURL := fmt.Sprintf("%s/%s/events/%s/%s", config.BackendURL, event.Entity.Namespace, event.Entity.Name, event.Check.Name)
	msgTitle := fmt.Sprintf("%s - %s%s - %s", msgThreadStatusValue, msgNamespace, event.Entity.Name, event.Check.Name)
	msgThreadTitle := fmt.Sprintf("%s%s - %s", msgNamespace, event.Entity.Name, event.Check.Name)
	msgExternalThreadID := fmt.Sprintf("%s%s-%s", msgNamespace, event.Entity.Name, event.Check.Name)
	msgThreadBody := fmt.Sprintf("%s", event.Check.Output)

	message := FlowdockMessage{
		Flowtoken: config.FlowdockToken,
		Event:     "activity",
		Author: FlowdockMessageAuthor{
			Name:   config.AuthorName,
			Avatar: config.AuthorAvatar,
		},
		Title:            msgTitle,
		ExternalThreadID: msgExternalThreadID,
		Thread: FlowdockMessageThread{
			Title: msgThreadTitle,
			Fields: []FlowdockMessageThreadFields{
				{Label: "Status", Value: msgThreadStatusValue},
			},
			Body:        msgThreadBody,
			ExternalURL: msgThreadExternalURL,
			Status: FlowdockMessageThreadStatus{
				Color: msgThreadStatusColor,
				Value: msgThreadStatusValue,
			},
		},
	}

	for k, v := range event.Entity.Labels {
		prefixstrip := len(labelPrefix)
		if strings.HasPrefix(k, labelPrefix) {
			tempfield := FlowdockMessageThreadFields{Label: k[prefixstrip:], Value: v}
			message.Thread.Fields = append(message.Thread.Fields, tempfield)
		}
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("Failed to marshal Flowdock message: %s", err)
	}

	resp, err := http.Post(config.FlowdockAPIURL, "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {
		return fmt.Errorf("Post to %s failed: %s", config.FlowdockAPIURL, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("POST to %s failed with %v", config.FlowdockAPIURL, resp.Status)
	}

	return nil
}
