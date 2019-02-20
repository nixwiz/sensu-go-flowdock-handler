# sensu-go-flowdock-handler
The Senso Go Flowdock Handler is a [Sensu Event Handler][1] for sending incident
notifications to CA Flowdock.

## Installation
Create an executable script from this source.

From the local path of the sensu-go-flowdock-handler repository:

```
go build -o /usr/local/bin/sensu-flowdock-handler main.go
```

[1]: https://docs.sensu.io/sensu-go/5.2/reference/handlers/#how-do-sensu-handlers-work

## Configuration

Example Sensu Go definition:

```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "flowdock"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-flowdock-handler -t 0123456789abcdef0123456789abcdef -b http://sensu-backend.example.com:3000",
        "timeout": 10,
        "filters": [
            "is_incident",
            "not_silenced"
        ]
    }
}

```

## Usage Examples

#### Help
```
The Sensu Flowdock handler for sending notifications

Usage:
  sensu-flowdock-handler [flags]

Flags:
  -a, --authorAvatar string    Avatar URL (default "https://avatars1.githubusercontent.com/u/1648901?s=200&v=4")
  -n, --authorName string      Name for the author of the thread (default "Sensu")
  -b, --backendURL string      The URL for the backend, used to create links to events
  -t, --flowdockToken string   The Flowdock application token
  -h, --help                   help for sensu-flowdock-handler
  -l, --labelPrefix string     Label prefix for entity fields to be included in thread (default "flowdock_")
```

#### Environment Variables
|Variable|Setting|
|--------------------|-------|
|FLOWDOCK_TOKEN| same as -t / --flowdockToken|
|FLOWDOCK_BACKENDURL|same as -b / --backendURL|
