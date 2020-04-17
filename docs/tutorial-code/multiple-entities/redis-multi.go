package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/integration"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
}

const (
	integrationName    = "com.myorganization.redis-multi"
	integrationVersion = "0.1.0"
	instanceOnePort    = 16379
	instanceTwoPort    = 26379
)

var (
	args argumentList
)

func queryGaugeRedisInfo(query string, port int) (float64, error) {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("redis-cli -p %d info | grep %s", port, query))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	splittedLine := strings.Split(string(output), ":")
	if len(splittedLine) != 2 {
		return 0, errors.New("cannot split the output line")
	}
	return strconv.ParseFloat(strings.TrimSpace(splittedLine[1]), 64)
}

func queryAttrRedisInfo(query string, port int) (string, string) {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("redis-cli -p %d info | grep %s", port, query))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", ""
	}
	splittedLine := strings.Split(string(output), ":")
	if len(splittedLine) != 2 {
		return "", ""
	}
	return strings.TrimSpace(splittedLine[0]), strings.TrimSpace(splittedLine[1])
}

func main() {
	// Create Integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	panicOnErr(err)

	// Create Entity, entities name must be unique
	e1, err := i.NewEntity("instance-1", "instance-1", "redis")
	panicOnErr(err)
	// Add event when redis starts
	if args.All() || args.Events {
		uptime, err := queryGaugeRedisInfo("uptime_in_seconds:", instanceOnePort)
		panicOnErr(err)
		if uptime < 60 {
			ev, _ := event.NewNotification("Redis Server recently started")
			e1.AddEvent(ev)
		}
		panicOnErr(err)
		if uptime < 60 {
			ev, _ := event.New(time.Now(), "summary", "category")
			e1.AddEvent(ev)
		}
		panicOnErr(err)
	}

	// Add Inventory item
	if args.All() || args.Inventory {
		key, value := queryAttrRedisInfo("redis_version", instanceOnePort)
		if key != "" {
			err = e1.AddInventoryItem(key, "value", value)
		}
		panicOnErr(err)
	}

	// Add Metric
	if args.All() || args.Metrics {
		metricValue, err := queryGaugeRedisInfo("instantaneous_ops_per_sec:", instanceOnePort)
		panicOnErr(err)
		g, _ := integration.Gauge(time.Now(), "query.instantaneousOpsPerSecond", metricValue)
		e1.AddMetric(g)
	}

	// Add the first entity to the integration
	i.AddEntity(e1)

	// Create another Entity
	e2, err := i.NewEntity("instance-2", "my-instance", "redis")
	panicOnErr(err)

	// Add event when redis starts
	if args.All() || args.Events {
		uptime, err := queryGaugeRedisInfo("uptime_in_seconds:", instanceOnePort)
		panicOnErr(err)
		if uptime < 60 {
			ev, _ := event.NewNotification("Redis Server recently started")
			e2.AddEvent(ev)
		}
		panicOnErr(err)
		if uptime < 60 {
			ev, _ := event.New(time.Now(), "summary", "category")
			e2.AddEvent(ev)
		}
		panicOnErr(err)
	}

	// Add Inventory item
	if args.All() || args.Inventory {
		key, value := queryAttrRedisInfo("redis_version", instanceTwoPort)
		if key != "" {
			err = e2.AddInventoryItem(key, "value", value)
		}
		panicOnErr(err)
	}

	if args.All() || args.Metrics {
		metricValue, err := queryGaugeRedisInfo("instantaneous_ops_per_sec:", instanceTwoPort)
		panicOnErr(err)
		g, _ := integration.Gauge(time.Now(), "query.instantaneousOpsPerSecond", metricValue)
		e2.AddMetric(g)
	}

	// Add the second entity to the integration
	i.AddEntity(e2)

	// Print the JSON document to stdout
	panicOnErr(i.Publish())
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
