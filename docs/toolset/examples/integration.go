package main

import (
	"os"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/sirupsen/logrus"
)

func main() {

	payload, err := integration.New("my-integration-data", "1.0")

	myHost := payload.LocalEntity()

	cpu, err := myHost.NewMetricSet("CpuSample")
	cpu.SetMetric("cpuPercent", 75.0, metric.GAUGE)

	disk, err := myHost.NewMetricSet("DiskSample")
	cpu.SetMetric("readsPerSecond", 12, metric.RATE)
	cpu.SetMetric("readBytes", 134, metric.DELTA)
	cpu.SetMetric("diskStatus", "OK!", metric.ATTRIBUTE)

}


