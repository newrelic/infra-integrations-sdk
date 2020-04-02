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
	Hostname string `default:"localhost" help:"Hostname or IP where Redis server is running."`
	Port     int    `default:"6379" help:"Port on which Redis server is listening."`
}

const (
	integrationName    = "com.myorg.redis"
	integrationVersion = "0.1.0"
)

var args argumentList

func queryRedisInfo(query string) (float64, error) {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("redis-cli -h %s -p %d info | grep %s", args.Hostname, args.Port, query))
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

func queryRedisConfig(query string) (string, string) {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("redis-cli -h %s -p %d CONFIG GET %s", args.Hostname, args.Port, query))

	output, err := cmd.CombinedOutput()
	panicOnErr(err)

	splittedLine := strings.Split(string(output), "\n")
	return splittedLine[0], splittedLine[1]
}

func main() {
	// Create Integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	panicOnErr(err)

	entity, err := i.NewEntity("redis_01", "my-redis", "RedisServer")
	panicOnErr(err)

	// Add Event
	if args.All() || args.Events {
		uptime, err := queryRedisInfo("uptime_in_seconds:")
		panicOnErr(err)
		if uptime < 60 {
			err = entity.AddEvent(event.NewNotification("Redis Server recently started"))
		}
		panicOnErr(err)
		if uptime < 60 {
			err = entity.AddEvent(event.New(time.Now(), "summary", "category"))
		}
		panicOnErr(err)
	}

	// Add Inventory item
	if args.All() || args.Inventory {
		key, value := queryRedisConfig("dbfilename")
		err = entity.AddInventoryItem(key, "value", value)
		panicOnErr(err)

		key, value = queryRedisConfig("bind")
		err = entity.AddInventoryItem(key, "value", value)
		panicOnErr(err)
	}

	// Add Metric
	if args.All() || args.Metrics {
		metricValue, err := queryRedisInfo("instantaneous_ops_per_sec:")
		panicOnErr(err)
		entity.AddMetric(integration.Gauge(time.Now(), "query.instantaneousOpsPerSecond", metricValue))
		//TODO do we support this?
		//metricValue1, err := queryRedisInfo("total_connections_received:")
		//panicOnErr(err)
		//err = entity.AddMetric("net.connectionsReceivedPerSecond", metricValue1, metric.RATE)
	}

	panicOnErr(i.Publish())
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
