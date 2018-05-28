package main

import (
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/integration"
)

func main() {

	payload, _ := integration.New("my-integration-data", "1.0")

	myHost := payload.LocalEntity()

	myHost.AddEvent(event.New("/etc/httpd.conf configuration file has changed", "config"))
	myHost.AddEvent(event.New("Service httpd has been restarted", "services"))

	payload.Publish()
}


