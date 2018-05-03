package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
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

func populateInventory(e *integration.Entity) error {
	cmd := exec.Command("/bin/sh", "-c", "redis-cli CONFIG GET dbfilename")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	splittedLine := strings.Split(string(output), "\n")
	if splittedLine[0] == "dbfilename" {
		e.SetInventoryItem(splittedLine[0], "value", splittedLine[1])
	}

	cmd = exec.Command("/bin/sh", "-c", "redis-cli CONFIG GET bind")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	splittedLine = strings.Split(string(output), "\n")
	if splittedLine[0] == "bind" {
		e.SetInventoryItem(splittedLine[0], "value", splittedLine[1])
	}
	return nil
}

func populateMetrics(ms *metric.Set) error {
	cmd := exec.Command("/bin/sh", "-c", "redis-cli info | grep instantaneous_ops_per_sec:")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	splittedLine := strings.Split(string(output), ":")
	if len(splittedLine) != 2 {
		return fmt.Errorf("cannot split the output line")
	}
	metricValue, err := strconv.ParseFloat(strings.TrimSpace(splittedLine[1]), 64)
	if err != nil {
		return err
	}
	ms.SetMetric("query.instantaneousOpsPerSecond", metricValue, metric.GAUGE) // nolint: errcheck

	cmd = exec.Command("/bin/sh", "-c", "redis-cli info | grep total_connections_received:")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	splittedLine = strings.Split(string(output), ":")
	if len(splittedLine) != 2 {
		return fmt.Errorf("cannot split the output line")
	}
	metricValue, err = strconv.ParseFloat(strings.TrimSpace(splittedLine[1]), 64)
	if err != nil {
		return err
	}

	ms.SetMetric("net.connectionsReceivedPerSecond", metricValue, metric.RATE) // nolint: errcheck
	return nil
}

func populateEvents(e *integration.Entity) error {
	cmd := exec.Command("/bin/sh", "-c", "redis-cli info | grep uptime_in_seconds:")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	splittedLine := strings.Split(string(output), ":")
	if len(splittedLine) != 2 {
		return fmt.Errorf("cannot split the output line")
	}
	uptime, err := strconv.ParseFloat(strings.TrimSpace(splittedLine[1]), 64)
	if err != nil {
		return err
	}
	if uptime < 60 {
		err = e.AddEvent(event.NewNotification("Redis Server recently started"))
		if err != nil {
			return err
		}
	}
	if uptime < 60 {
		err = e.AddEvent(event.New("Redis Server recently started", "redis-server"))
	}

	return err
}

func main() {
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	panicOnErr(err)

	e := i.LocalEntity()

	if args.All || args.Inventory {
		panicOnErr(populateInventory(e))
	}

	if args.All || args.Metrics {
		ms, err := e.NewMetricSet("RedisSample")
		panicOnErr(err)
		panicOnErr(populateMetrics(ms))
	}

	if args.All || args.Events {
		err := populateEvents(e)
		if err != nil {
			i.Logger().Debugf("adding event failed, got: %s", err)
		}
	}
	panicOnErr(i.Publish())
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
