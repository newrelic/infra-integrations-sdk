# Integration data model

For detailed information about the `integration` package, please refer to the
[Integration package GoDoc page](https://godoc.org/github.com/newrelic/infra-integrations-sdk/integration)

## Building an integration

To build an integration with the default configuration, you have to invoke the
[integration.New](https://godoc.org/github.com/newrelic/infra-integrations-sdk/integration#New) function with
the integration name and version as argument:

```go
payload, err := integration.New("my-integration", "1.0")
```

The above invocation will return an integration with the default configuration:

* The final integration JSON payload is sent to the standard output.
* The [logging](log.md) messages are submitted to the standard error (with `INFO` level).
* No synchronization capabilities are set. That means that you cannot safely add data from different, concurrent
  threads.
* A persistent [Storer](persist.md) is set, whose contents will be stored in a file whose path can be constructed as
  `<OS temp dir>/nr-integrations/<integration name>.json`, whith a default 1-minute _Time To Live_.
* Configuration specified in the [default arguments](args.md).

The `integration.New` function accepts, as a variable number of arguments, diverse configuration options. For example,
the following code would create an integration which logs data to a [Logrus](https://github.com/sirupsen/logrus)
logger implementation, is synchronized so data can be managed from concurrent threads, and writes the JSON output
to a file called `output.json` instead of the default standard output.

```go
payloadFile, _ := os.Create("output.json")

payload, err := integration.New("my-integration-data", "1.0",
        integration.Logger(logrus.New()),
        integration.Synchronized(),
        integration.Writer(payloadFile),
    )
```

For more details, check the documentation of the
[Option interface implementations](https://godoc.org/github.com/newrelic/infra-integrations-sdk/integration#Option).

## Integration structure elements

### Entity

An entity represents a monitoring target (e.g. a host). Since GoSDK v3, a single JSON payload can handle data from
multiple entities (for example, the local host that runs the Infrastructure Agent + some remote or virtual hosts that
are being monitored by the integration).

To instantiate a new entity, you can use the `LocalEntity` or `Entity` functions from the 
[Integration type](https://godoc.org/github.com/newrelic/infra-integrations-sdk/integration#Integration):

```go
payload, err := integration.New("my-integration-data", "1.0")

// Creates a local entity
localhost := payload.LocalEntity()

// Creates a "remote" entity, given an entity name and a namespace
container, err := payload.Entity("my-cloud-resource", "my-namespace")
```

Each entity has three sections: `metrics`, `inventory`, and `events`, which will be explained in the following
subsections.

### Metrics

Metrics are quantifiable measurements associated to a given entity (e.g. percentage of CPU, or requests/second). There
are four type of metrics:

* `GAUGE`: an absolute, spot value, such as percentage of used CPU or free system memory.
* `RATE`: a value that represents a measured quantity on a time period. For example, transferred bytes per second.
* `DELTA`: a value that represents the variation of a quantity since the last sample. For example, number of new
  connections since the last metric sample was set.
* `ATTRIBUTE`: a string value (e.g. `"high"` or `"stopped"`)

Metrics are grouped in a key-value map (called _metrics set_). Every entity can control several, different metric sets.
The `Entity` type provides functions to create metric sets (as well as for inventory and events):

```go
myHost := payload.LocalEntity()

cpu, err := myHost.NewMetricSet("CpuSample")
cpu.SetMetric("cpuPercent", 75.0, metric.GAUGE)

disk, err := myHost.NewMetricSet("DiskSample")
cpu.SetMetric("readsPerSecond", 12, metric.RATE)
cpu.SetMetric("readBytes", 134, metric.DELTA)
cpu.SetMetric("diskStatus", "OK!", metric.ATTRIBUTE)
```

The above example creates two metric sets for a same entity. The `NewMetricSet` function accepts the name of the
metric set as a parameter, then the `SetMetric` function requires the name, value as well as the type of the metric.

Please refer to the [Metrics GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/data/metric) for a
detailed description of the metrics API.

### Inventory

Inventory provides track of a set of available items, as well as some associated data to them. For example, the
installed software components, with their version number.



Please refer to the [Inventory GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/data/inventory) for a
detailed description of the inventory API.

### Events

Please refer to the [Events GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/data/event) for a
detailed description of the events API.

