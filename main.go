package main

// To do:
// check return from post to return to calling func
// Add additional status Fields
// use labels/annotations for templating (extra key/value pair(s))
// reconcile usage of log versus other error output and use of panic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

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

var (
	flowdockToken string
	authorName    string
	authorAvatar  string
	backendURL    string
	stdin         *os.File

	threadBody                  string
	msgTitle                    string
	msgThreadTitle              string
	msgThreadExternalURL        string
	msgThreadStatusColor        string
	msgThreadStatusValue        string
	msgExternalThreadIdTemplate = "{{.Entity.System.Hostname}}"
	msgThreadBodyTemplate       = "{{.Check.Output}}"
)

func main() {

	cmd := &cobra.Command{
		Use:   "sensu-flowdock-handler",
		Short: "The Sensu Flowdock handler for sending notifications",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&flowdockToken, "flowdockToken", "t", "", "The Flowdock Application API key")
	cmd.Flags().StringVarP(&authorName, "authorName", "n", "Sensu", "The name to list as author of the thread")
	cmd.Flags().StringVarP(&authorAvatar, "authorAvatar", "a", "https://avatars1.githubusercontent.com/u/1648901?s=200&v=4", "The URL for the avatar to use for the thread")
	cmd.Flags().StringVarP(&backendURL, "backendURL", "b", "", "The URL for the backend server used to create links to events")
	cmd.Execute()

}

func run(cmd *cobra.Command, args []string) error {

	validationError := checkArgs()
	if validationError != nil {
		return validationError
	}

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

	flowDockError := sendFlowDock(event)
	if flowDockError != nil {
		return fmt.Errorf("failed to send to Flowdock: %s", flowDockError)
	}

	return nil

}

func checkArgs() error {

	if len(flowdockToken) == 0 {
		return errors.New("missing Flowdock token")
	}
	if len(authorName) == 0 {
		return errors.New("missing author name supecification")
	}
	if len(authorAvatar) == 0 {
		return errors.New("missing author avatar URL specification")
	}
	if len(backendURL) == 0 {
		return errors.New("missing backend URL specification")
	}
	backendURL = strings.TrimSuffix(backendURL, "/")

	return nil
}

func sendFlowDock(event *types.Event) error {

	var (
		msgThreadStatusColor string
		msgThreadStatusValue string
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

	msgThreadExternalURL := fmt.Sprintf("%s/%s/events/%s/%s", backendURL, event.Entity.Namespace, event.Entity.Name, event.Check.Name)
	msgTitle := fmt.Sprintf("%s - %s - %s", msgThreadStatusValue, event.Entity.Name, event.Check.Name)
	msgThreadTitle := fmt.Sprintf("%s - %s - %s", event.Entity.Name, event.Check.Name, msgThreadStatusValue)

	msgExternalThreadId, err := resolveTemplate(msgExternalThreadIdTemplate, event)
	if err != nil {
		return err
	}
	msgThreadBody, err := resolveTemplate(msgThreadBodyTemplate, event)
	if err != nil {
		return err
	}

	message := FlowDockMessage{
		Flowtoken: flowdockToken,
		Event:     "activity",
		Author: FlowDockMessageAuthor{
			Name:   authorName,
			Avatar: authorAvatar,
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

	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post("https://api.flowdock.com/messages", "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	log.Println(result)
	log.Println(result["data"])

	return nil
}

func resolveTemplate(templateValue string, event *types.Event) (string, error) {

	var resolved bytes.Buffer
	tmpl, err := template.New("test").Parse(templateValue)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(&resolved, *event)
	if err != nil {
		panic(err)
	}

	return resolved.String(), nil

}