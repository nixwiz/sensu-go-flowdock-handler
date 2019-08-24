# sensu-go-flowdock-handler
The Senso Go Flowdock Handler is a [Sensu Event Handler][1] for sending incident
notifications to CA Flowdock.

## Installation
Create an executable binary from this source, create an [asset][7] from [bonsai][6], or download one of the existing [releases][5].

From the local path of the sensu-go-flowdock-handler repository:

```
go build -o /usr/local/bin/sensu-go-flowdock-handler main.go
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
        "command": "sensu-go-flowdock-handler -t 0123456789abcdef0123456789abcdef -b http://sensu-backend.example.com:3000",
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
The Sensu Go Flowdock handler for sending notifications

Usage:
  sensu-go-flowdock-handler [flags]

Flags:
  -a, --autherAvatar string    Avatar URL (default "https://avatars1.githubusercontent.com/u/1648901?s=200&v=4")
  -n, --authorName string      Name for the auther of the thread (default "Sensu")
  -b, --backendURL string      The URL for the backend, used to create links to events
  -t, --flowdockToken string   The Flowdock application token
  -h, --help                   help for sensu-go-flowdock-handler
  -i, --includeNamespace       Include the namespace with the entity name in title and thread ID
  -l, --labelPrefix string     Label prefix for entity fields to be included in thread
```

#### Environment Variables and Annotations
|Environment Variable|Setting|Annotation|
|--------------------|-------|----------|
|SENSU_FLOWDOCK_TOKEN| same as -t / --flowdockToken|sensu.io/plugins/flowdock/flowdockToken|
|SENSU_FLOWDOCK_BACKENDURL|same as -b / --backendURL|sensu.io/plugins/flowdock/backendURL|
|N/A|same as -n / --authorName|sensu.io/plugins/flowdock/authorName|
|N/A|same as -a / --authorAvatar|sensu.io/plugins/flowdock/authorAvatar|

#### Precedence
environment variable < command-line argument < annotation

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

[1]: https://docs.sensu.io/sensu-go/latest/reference/handlers/#how-do-sensu-handlers-work
[2]: https://www.flowdock.com/api/integration-getting-started
[3]: https://www.flowdock.com/oauth/applications
[4]: https://www.flowdock.com/api/sources
[5]: https://github.com/nixwiz/sensu-go-flowdock-handler/releases
[6]: https://bonsai.sensu.io/assets/nixwiz/sensu-go-flowdock-handler
[7]: https://docs.sensu.io/sensu-go/latest/reference/assets/
