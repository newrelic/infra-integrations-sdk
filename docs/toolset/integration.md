# Integration data model

The integration JSON payload contains the complete information sample that an integration sends through the New Relic
Infrastructure Agent, at a given moment.

This document describes the basic structure and elements of the integration payload, as well as the basic API by means
of simple examples.

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
* A persistent [Storer](persist.md) is set, whose contents will be stored in a file whose path can be constructed as
  `<OS temp dir>/nr-integrations/<integration name>.json`, whith a default 1-minute _Time To Live_.
* Configuration specified in the [default arguments](args.md).

The `integration.New` function accepts, as a variable number of arguments, diverse configuration options. For example,
the following code would create an integration which logs data to a [Logrus](https://github.com/sirupsen/logrus)
logger implementation. In addition, the
output JSON payload is written to a file called `output.json` instead of the default standard output.

```go
payloadFile, _ := os.Create("output.json")

payload, err := integration.New("my-integration-data", "1.0",
        integration.Logger(logrus.New()),
        integration.Writer(payloadFile),
    )
```

For more details, check the documentation of the
[Option interface implementations](https://godoc.org/github.com/newrelic/infra-integrations-sdk/integration#Option).

You can safely add data from different concurrent threads since the sdk is thread safe.

## Integration structure elements

An integration JSON payload contains data from multiple entities. Each `entity` stores information about `metrics`,
`inventory` and `events`.

The rest of this section describes all the concepts that are part of the integration JSON payload, as well as their
basic composition through the `GoSDK v3` API.

### Entity

An entity represents a monitoring target (e.g. a host). Since `GoSDK v3`, a single JSON payload can handle data from
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
* `PRATE`: version of `RATE` that only allows positive values.
* `DELTA`: a value that represents the variation of a quantity since the last sample. For example, number of new
  connections since the last metric sample was set.
* `PDELTA`: version of `DELTA` that only allows positive values.
* `ATTRIBUTE`: a string value (e.g. `"high"` or `"stopped"`)

Metrics are grouped in a key-value map (called _metrics set_). Every entity can control several, different metric sets.
The `Entity` type provides functions to create metric sets (as well as for inventory and events):

```go
myHost := payload.LocalEntity()

cpu := myHost.NewMetricSet("CpuSample")
err = cpu.SetMetric("cpuPercent", 75.0, metric.GAUGE)
```

**RATE, PRATE, DELTA and PDELTA require to belong to at least 1 attribute.**

As they are flushed to disk this attribute is used to "namespace" the metrics on the set so they don't collide with others with the same name.

So `NewMetricSet` provides an optional list of `metric.Attribute` arguments. There's a constructor function that comes handy to create one: `metric.Attr`.

The attributes provided on the `NewMetricSet` constructor are also added as usual attribute metrics.

If no `Attribute` is provided to `NewMetricSet`, an `error` value will be returned when calling `SetMetric` for a RATE, PRATE, DELTA or PDELTA.

```go
disk := myHost.NewMetricSet("DiskSample", metric.Attr("diskStatus", "OK"))
err1 = disk.SetMetric("readsPerSecond", 12, metric.RATE)
err2 = disk.SetMetric("readBytes", 134, metric.DELTA)
```

The above example creates two metric sets for a same entity. The `NewMetricSet` function accepts the name of the
metric set as a parameter, then the `SetMetric` function requires the name, value as well as the type of the metric.

Please refer to the [Metrics GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/data/metric) for a
detailed description of the metrics API.

### Inventory

Inventory provides track of a set of available items, as well as some associated data to them. For example, the
installed software components, with their version number.

Inventory items are stored as a 2-level map, where the first-level key represents the inventory item and the 
second-level key plus its associated value represents a given aspect of such inventory item.

To set the inventory items and associated values, the `Entity` type provides the
`SetInventoryItem(<item name>, <key name>, <value>)`. For example, the following code keeps track of the items inside
a remote entity named _"fridge"_ in the _"beach-home"_ namespace, and describes different characteristics for some
of the items inside the fridge:

```go
payload, _ := integration.New("my-integration-data", "1.0")

food, _ := payload.Entity("fridge", "beach-home")

food.SetInventoryItem("beans", "weight", 125.0)
food.SetInventoryItem("beans", "brand", "Heinz")
food.SetInventoryItem("beer", "size", 1000)
food.SetInventoryItem("beer", "type", "dark")
food.SetInventoryItem("coffee", "decaf", "yes")

payload.Publish()
``` 

The above code would insert the next inventory data into the integration JSON payload:

```json
"inventory": {
    "beans": {
      "brand": "Heinz",
      "weight": 125
    },
    "beer": {
      "size": 1000,
      "type": "dark"
    },
    "coffee": {
      "decaf": "yes"
    }
  },
```

Please refer to the [Inventory GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/data/inventory) for a
detailed description of the inventory API.

### Events

Events describe meaningful things that happen at a given moment (e.g. a host has been started, a package has been
removed, a configuration property has changed...).

An event has two fields: `summary` and `category`. Category is a simple keyword to group events. Summary is a
human-readable descriptive message. An event is created by the `event.New` function, and added by means of the
`AddEvent` function of the `Entity` type.

For example, the following code would add two different events from different categories:

```go
myHost := payload.LocalEntity()

myHost.AddEvent(event.New("/etc/httpd.conf configuration file has changed", "config"))
myHost.AddEvent(event.New("Service httpd has been restarted", "services"))
```

Please refer to the [Events GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/data/event) for a
detailed description of the events API.

