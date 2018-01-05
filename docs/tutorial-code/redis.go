package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
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

func populateInventory(inventory sdk.Inventory) error {
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

func populateMetrics(ms *metric.MetricSet) error {
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
	ms.SetMetric("query.instantaneousOpsPerSecond", metricValue, metric.GAUGE)

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
	ms.SetMetric("net.connectionsReceivedPerSecond", metricValue, metric.RATE)

	return nil
}

func populateEvents(integration *sdk.Integration) error {
	var err error
	for _, m := range integration.Metrics {
		mVal := *m
		if v, ok := mVal["net.connectionsReceivedPerSecond"]; ok {
			if v.(float64) > 1000 {
				err = integration.AddEvent("More than 1000 connections received per second", "notification")
			}
		}
	}
	return err
}

func main() {
	integration, err := sdk.NewIntegration(integrationName, integrationVersion, &args)
	fatalIfErr(err)

	if args.All || args.Inventory {
		fatalIfErr(populateInventory(integration.Inventory))
	}

	if args.All || args.Metrics {
		ms := integration.NewMetricSet("RedisSample")
		fatalIfErr(populateMetrics(ms))

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
