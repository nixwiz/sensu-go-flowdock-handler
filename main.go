package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

type FlowDockMessageAuthor struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type FlowDockMessageThreadStatus struct {
	Color string `json:"color"`
	Value string `json:"value"`
}

type FlowDockMessageThreadFields struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type FlowDockMessageThread struct {
	Title        string                        `json:"title"`
	Fields       []FlowDockMessageThreadFields `json:"fields"`
	Body         string                        `json:"body"`
	External_url string                        `json:"external_url"`
	Status       FlowDockMessageThreadStatus   `json:"status"`
}

type FlowDockMessage struct {
	Flowtoken          string                `json:"flow_token"`
	Event              string                `json:"event"`
	Author             FlowDockMessageAuthor `json:"author"`
	Title              string                `json:"title"`
	External_thread_id string                `json:"external_thread_id"`
	Thread             FlowDockMessageThread `json:"thread"`
}

type HandlerConfigOption struct {
	Value string
	Path  string
	Env   string
}

type HandlerConfig struct {
	FlowdockToken HandlerConfigOption
	AuthorName    HandlerConfigOption
	AuthorAvatar  HandlerConfigOption
	BackendURL    HandlerConfigOption
	Keyspace      string
}

const flowdockAPIURL string = "https://api.flowdock.com/messages"

var (
	labelPrefix      string
	includeNamespace bool
	stdin            *os.File

	threadBody           string
	msgTitle             string
	msgThreadTitle       string
	msgThreadExternalURL string
	msgThreadStatusColor string
	msgThreadStatusValue string

	config = HandlerConfig{
		FlowdockToken: HandlerConfigOption{Path: "token", Env: "SENSU_FLOWDOCK_TOKEN"},
		AuthorName:    HandlerConfigOption{Value: "Sensu", Path: "author-name"},
		AuthorAvatar:  HandlerConfigOption{Value: "https://avatars1.githubusercontent.com/u/1648901?s=200&v=4", Path: "author-avatar"},
		BackendURL:    HandlerConfigOption{Path: "backend-url", Env: "SENSU_FLOWDOCK_BACKENDURL"},
		Keyspace:      "sensu.io/plugins/flowdock/config",
	}
	options = []*HandlerConfigOption{
		&config.FlowdockToken,
		&config.AuthorName,
		&config.AuthorAvatar,
		&config.AuthorAvatar,
	}
)

func main() {

	cmd := &cobra.Command{
		Use:   "sensu-flowdock-handler",
		Short: "The Sensu Flowdock handler for sending notifications",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&config.FlowdockToken.Value, "flowdockToken", "t", os.Getenv(config.FlowdockToken.Env), "The Flowdock application token, if not in env "+config.FlowdockToken.Env)
	cmd.Flags().StringVarP(&config.AuthorName.Value, "authorName", "n", config.AuthorName.Value, "Name for the author of the thread")
	cmd.Flags().StringVarP(&config.AuthorAvatar.Value, "authorAvatar", "a", config.AuthorAvatar.Value, "Avatar URL")
	cmd.Flags().StringVarP(&config.BackendURL.Value, "backendURL", "b", os.Getenv(config.BackendURL.Env), "The URL for the backend, used to create links to events, if not in env "+config.BackendURL.Env)
	cmd.Flags().StringVarP(&labelPrefix, "labelPrefix", "l", "flowdock_", "Label prefix for entity fields to be included in thread")
	cmd.Flags().BoolVarP(&includeNamespace, "includeNamespace", "i", false, "Include the namespace with the entity name in title and thread ID")
	cmd.Execute()

}

func run(cmd *cobra.Command, args []string) error {

	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err)
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", err)
	}

	if err = event.Validate(); err != nil {
		return fmt.Errorf("failed to validate event: %s", err)
	}

	if !event.HasCheck() {
		return fmt.Errorf("event does not contain check")
	}

	configurationOverrides(&config, options, event)

	validationError := checkArgs()
	if validationError != nil {
		return validationError
	}

	flowDockError := sendFlowDock(event)
	if flowDockError != nil {
		return fmt.Errorf("failed to send to Flowdock: %s", flowDockError)
	}

	return nil

}

func checkArgs() error {

	if len(config.FlowdockToken.Value) == 0 {
		return errors.New("missing Flowdock token")
	}
	if len(config.AuthorName.Value) == 0 {
		return errors.New("missing author name supecification")
	}
	if len(config.AuthorAvatar.Value) == 0 {
		return errors.New("missing author avatar URL specification")
	}
	if len(config.BackendURL.Value) == 0 {
		return errors.New("missing backend URL specification")
	}
	if !govalidator.IsURL(config.BackendURL.Value) {
		return errors.New("invlaid backend URL specification")
	}
	config.BackendURL.Value = strings.TrimSuffix(config.BackendURL.Value, "/")

	return nil
}

func sendFlowDock(event *types.Event) error {

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

	if includeNamespace {
		msgNamespace = event.Entity.Namespace + "/"
	} else {
		msgNamespace = ""
	}

	msgThreadExternalURL := fmt.Sprintf("%s/%s/events/%s/%s", config.BackendURL.Value, event.Entity.Namespace, event.Entity.Name, event.Check.Name)
	msgTitle := fmt.Sprintf("%s - %s%s - %s", msgThreadStatusValue, msgNamespace, event.Entity.Name, event.Check.Name)
	msgThreadTitle := fmt.Sprintf("%s%s - %s", msgNamespace, event.Entity.Name, event.Check.Name)
	msgExternalThreadId := fmt.Sprintf("%s%s-%s", msgNamespace, event.Entity.Name, event.Check.Name)
	msgThreadBody := fmt.Sprintf("%s", event.Check.Output)

	message := FlowDockMessage{
		Flowtoken: config.FlowdockToken.Value,
		Event:     "activity",
		Author: FlowDockMessageAuthor{
			Name:   config.AuthorName.Value,
			Avatar: config.AuthorAvatar.Value,
		},
		Title:              msgTitle,
		External_thread_id: msgExternalThreadId,
		Thread: FlowDockMessageThread{
			Title: msgThreadTitle,
			Fields: []FlowDockMessageThreadFields{
				{Label: "Status", Value: msgThreadStatusValue},
			},
			Body:         msgThreadBody,
			External_url: msgThreadExternalURL,
			Status: FlowDockMessageThreadStatus{
				Color: msgThreadStatusColor,
				Value: msgThreadStatusValue,
			},
		},
	}

	for k, v := range event.Entity.Labels {
		prefixstrip := len(labelPrefix)
		if strings.HasPrefix(k, labelPrefix) {
			tempfield := FlowDockMessageThreadFields{Label: k[prefixstrip:], Value: v}
			message.Thread.Fields = append(message.Thread.Fields, tempfield)
		}
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("Failed to marshal Flowdock message: %s", err)
	}

	resp, err := http.Post(flowdockAPIURL, "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {
		return fmt.Errorf("Post to %s failed: %s", flowdockAPIURL, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("POST to %s failed with %v", flowdockAPIURL, resp.Status)
	}

	return nil
}

func configurationOverrides(config *HandlerConfig, options []*HandlerConfigOption, event *types.Event) {
	if config.Keyspace == "" {
		return
	}
	for _, opt := range options {
		if opt.Path != "" {
			// compile the Annotation keyspace to look for configuration overrides
			k := path.Join(config.Keyspace, opt.Path)
			switch {
			case event.Check.Annotations[k] != "":
				opt.Value = event.Check.Annotations[k]
				log.Printf("Overriding default handler configuration with value of \"Check.Annotations.%s\" (\"%s\")\n", k, event.Check.Annotations[k])
			case event.Entity.Annotations[k] != "":
				opt.Value = event.Entity.Annotations[k]
				log.Printf("Overriding default handler configuration with value of \"Entity.Annotations.%s\" (\"%s\")\n", k, event.Entity.Annotations[k])
			}
		}
	}
}
