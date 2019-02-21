# sensu-go-flowdock-handler
The Senso Go Flowdock Handler is a [Sensu Event Handler][1] for sending incident
notifications to CA Flowdock.

## Installation
Create an executable script from this source or download one of the existing [releases][5].

From the local path of the sensu-go-flowdock-handler repository:

```
go build -o /usr/local/bin/sensu-flowdock-handler main.go
```

## Sensu Configuration

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

## Flowdock Configuration

This handler makes use of Flowdock's "new" [Integration API][2] mechanism.  This means creating a [developer application][3]
and then a [source][4].  This source will have the API Token needed by this handler.

**Note:**  Actions for these messages are not implemented.

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

**Note:**  The command line arguments take precedence over the environment variables above.

#### Usage of entity labels to add fields to output
This handler can make use of labels provided by the entity to populate addtional fields in the thread.
```yaml
---
##
# agent configuration
##
#name: ""
namespace: "default"
subscriptions:
  - linux
backend-url:
  - "ws://127.0.0.1:8081"

labels:
  flowdock_Application: "webapp1"
  flowdock_Environment: "live"
```

## Sample in Flowdock
Below is a thread sample that includes surfacing the labels defined above.

![Flowdock Sample](https://toddcampbell.net/images/sensu_flowdock.png)

[1]: https://docs.sensu.io/sensu-go/5.2/reference/handlers/#how-do-sensu-handlers-work
[2]: https://www.flowdock.com/api/integration-getting-started
[3]: https://www.flowdock.com/oauth/applications
[4]: https://www.flowdock.com/api/sources
[4]: https://github.com/nixwiz/sensu-go-flowdock-handler/releases
