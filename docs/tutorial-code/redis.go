package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk/v1"
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

func populateInventory(inventory v1.Inventory) error {
	cmd := exec.Command("/bin/sh", "-c", "redis-cli CONFIG GET dbfilename")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	splittedLine := strings.Split(string(output), "\n")
	if splittedLine[0] == "dbfilename" {
		inventory.SetItem(splittedLine[0], "value", splittedLine[1])
	}

	cmd = exec.Command("/bin/sh", "-c", "redis-cli CONFIG GET bind")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	splittedLine = strings.Split(string(output), "\n")
	if splittedLine[0] == "bind" {
		inventory.SetItem(splittedLine[0], "value", splittedLine[1])
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
		return fmt.Errorf("Cannot split the output line")
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
		return fmt.Errorf("Cannot split the output line")
	}
	metricValue, err = strconv.ParseFloat(strings.TrimSpace(splittedLine[1]), 64)
	if err != nil {
		return err
	}

	ms.SetMetric("net.connectionsReceivedPerSecond", metricValue, metric.RATE) // nolint: errcheck
	return nil
}

func populateEvents(integration *v1.Integration) error {
	cmd := exec.Command("/bin/sh", "-c", "redis-cli info | grep uptime_in_seconds:")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	splittedLine := strings.Split(string(output), ":")
	if len(splittedLine) != 2 {
		return fmt.Errorf("Cannot split the output line")
	}
	uptime, err := strconv.ParseFloat(strings.TrimSpace(splittedLine[1]), 64)
	if err != nil {
		return err
	}
	if uptime < 60 {
		err = integration.AddNotificationEvent("Redis Server recently started")
		if err != nil {
			return err
		}
	}
	if uptime < 60 {
		err = integration.AddEvent(v1.Event{Summary: "Redis Server recently started", Category: "redis-server"})
	}

	return err
}

func main() {
	integration, err := v1.NewIntegration(integrationName, integrationVersion, &args)
	fatalIfErr(err)

	if args.All || args.Inventory {
		fatalIfErr(populateInventory(integration.Inventory))
	}

	if args.All || args.Metrics {
		ms := integration.NewMetricSet("RedisSample")
		fatalIfErr(populateMetrics(&ms))
	}

	if args.All || args.Events {
		err := populateEvents(integration)
		if err != nil {
			log.Debug("adding event failed, got: %s", err)
		}
	}
	fatalIfErr(integration.Publish())
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
