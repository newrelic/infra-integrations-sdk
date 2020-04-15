# Upgrading from GoSDK v3.x to GoSDK v4.x

Version 4 of the GoSDK is not backward compatible with the previous versions. This document describes the changes 
between versions and how to update an integration built with the GoSDK v3.x.

The Go SDK v4 contains the following changes:

* New Infrastructure Agent Integration JSON version 4.
* Add support for dimensional metrics using the [Metrics API format][1].
* New metric data types: `count`, `summary`, `cumulative-count` and `cumulative-rate`.
* LocalEntity has been replaced by HostEntity.
* Support for Go Modules
* Removed support for protocols prior to v4.x.

## JSON schema changes

The following section explains the new JSON schema. Be aware that the new SDK only supports this new protocol version. 
These are the most important changes:

* A new "integration" object at the top-level.

* The "entity" and "metrics" objects have been modified.

### JSON protocol v4 sample
```
{
  "protocol_version":"4",                      # protocol version number
  "integration":{                              # this data will be added to all metrics and events as attributes,                                               
                                               # and also sent as inventory
    "name":"integration name",
    "version":"integration version"
  },
  "data":[                                    # List of objects containing entities, metrics, events and inventory
    {
      "entity":{                              # this object is optional. If it's not provided, then the Entity will get 
                                              # the same entity ID as the agent that executes the integration. 
        "name":"redis:192.168.100.200:1234",  # unique entity name per customer account
        "type":"RedisInstance",               # entity's category
        "displayName":"my redis instance",    # human readable name
        "metadata":{}                         # can hold general metadata or tags. Both are key-value pairs that will 
                                              # be also added as attributes to all metrics and events
      },
      "metrics":[                             # list of metrics using the dimensional metric format
        {
          "name":"redis.metric1",
          "type":"count",                     # gauge, count, summary, cumulative-count, rate or cumulative-rate
          "value":93, 
          "attributes":{}                     # set of key-value pairs that define the dimensions of the metric
        }
      ],
      "inventory":{...},                      # Inventory remains the same
      "events":[...]                          # Events remain the same
    }
  ]
}
```

## GoSDK v4 API changes

This section enumerates the main changes you have to keep in mind to upgrade from GoSDK v3.x to v4.x.

### Entities

In previous versions of the SDK, we had a distinction between the Local Entity and Remote entities. You could get the 
`Local Entity` calling the method

`func (i *Integration) LocalEntity() *Entity`

or create a `Remote Entity` calling

`func (i *Integration) Entity(name, namespace string, idAttributes ...IDAttribute) (e *Entity, err error)`

In SDK v4, you can obtain the Host Entity or create Entities. The Host Entity is a the entity attached to the host 
where the agent is running. It inherits the same entity ID as the agent which executes the integration. On the other
hand, an Entity has its own entity ID. Both types of entities are decorated with the metadata from the host where the
agent is running.

To get the Host Entity you need to access the `HostEntity` of an integration. Example:

```
// create the integration
i, err := integration.New(integrationName, integrationVersion)

// add some inventory to the Host Entity
i.HostEntity.AddInventoryItem("key", "value", "foo")
```

To create an Entity you need to call the following method inside  integration

`func (i *Integration) NewEntity(name, type, displayName string) (e *Entity, err error)`

Example: 

```
// create the integration
i, err := integration.New(integrationName, integrationVersion)

// create an entity
entity, err := i.NewEntity(entityName, entityType, entityDisplayName)
```

These are the parameters accepted by `NewEntity`:

* `entityName` must be an unique value per customer account because it uniquely identifies your Entity. It cannot be empty and 
  the SDK cannot validate nor enforce that uniqueness, so it's up to the client to define a naming schema that produces 
  unique names.

  For example, using the following schema `<service_name>:<ip|hostname>:<port>`, we could create these names: 
  `redis:192.168.100.200:12345` or `mysql:my_server_hostname:8086`

  For the ip or the hostname, please avoid using the values `127.0.0.1` or `localhost` as they don't produce unique 
  identifiers.

* `entityType` describes the Entity's category. It cannot be empty. Some examples: "RedisInstance", "DockerContainer", 
  "K8SPod", "CassandraCluster", "MySQLInstance", "KafkaBroker", "NginxServer", etc.

* `entityDisplayName` is a friendly human readable name to be used in the UI. It can be empty.

#### Migrating from LocalEntity

The LocalEntity must be replaced by the Host Entity. The Host Entity has the same methods and functionality as a normal
Entity. 

#### Adding metadata and tags to an Entity

Entities can have a list of metadata. This metadata can be used to search and filter your entities and will be added to 
all the metrics and events of the Entity.

There are two types of metadata:
* Generic metadata about the entity. The metadata is provided by the integration. 
* Tags. This information is provided by the user through the agent when it executes the integration. The SDK will prefix
automatically the key values with "tags.".

In this example we add the service version that produces the entity as generic metadata and the team that owns the data 
produced by the integration as a tag.

```
  ...
	"entity": {
		"name": "redis:192.168.100.200:1234",
		"type": "RedisInstance",
		"displayName": "my redis instance",
		"metadata": {
			"redis_version": "4.0",
			"tags.team": "foobar"
		}
	}
  ...
```

To add metadata just call 

`func (i *Entity) AddMetadata(key string, value interface{})`

To add a tag just call 

`func (i *Entity) AddTag(key string, value interface{})`

On both cases, if the `key` already exists the previous value will be overwritten with the new one.

### Metrics

#### Types

In this SDK v4 we introduce some new metric types while removing some others.

## Metric types in v4

* Gauge

  It is a value that can increase or decrease. It generally represents the value for something at a particular moment 
  in time.
  
  Examples include the CPU load or the memory consumption.

* Count

  Measures the number of occurrences of an event given a time interval.

  Examples include cache hits per reporting interval and the number of threads created per reporting interval.

* Summary

  Used to report pre-aggregated data, or information on aggregated discrete events. A summary 
  includes a `count`, `sum` value, `min` value, and `max` value. 

  Examples include transaction count/durations and queue count/ durations.
  
* Cumulative-count
  
  It generates a count metric by calculating the delta between the reported value and the previous one. Interval is 
  calculated based on the timestamp of the datapoints. The value passed in is not a delta but instead the absolute value. 
  
  Examples include the total number of threads created for a process or the total number of cache hits since process 
  startup.

* Rate
  
  It generates a gauge metric by dividing the current value by the interval in seconds. Interval is calculated based on the 
  timestamp of the datapoints.
  
  Examples include the number of get request per second.
  
* Cumulative-rate  

  Similar to the rate, but this time we calculate the delta between the reported value and the previous one and then we
  divide the result by the interval in seconds. Interval is calculated based on the timestamp of the datapoints.
  
  Examples include the total write bytes per second.

#### Creating v4 metrics

| metric            | method                                                                 |
|-------------------|------------------------------------------------------------------------|
| gauge             | integration.Gauge(timestamp, name, value)                              |
| count             | integration.Count(timestamp, name, count)                              |
| summary           | integration.Summary(timestamp, name, count, average, sum, min, max)    |
| cumulative-count  | integration.CumulativeCount(timestamp, name, count)                    |
| rate              | integration.Rate(timestamp, name, value)                               |
| cumulative-rate   | integration.CumulativeRate(timestamp, name, value)                     |

Description of the paramaters:

`timestamp` (time.Time) is the metric's start time in Unix time (milliseconds).

`name` (string) is the name of the metric. The value must be less than 255 characters.

`value` (float64) is a double. Gauges values can be any value positive or negative. Cumulative-rate values need to be 
ever-growing.

`count` (float64) is the number of occurrences of an event reported in a given interval. Must be a positive value.

`average` (float64) is the number expressing the central value in the set of data.

`sum` (float64) is the aggregation of all the values registered in the set of data.

`min` (float64) is the minimum value registered in the set of data.

`max` (float64) is the maximum value registered in the set of data.

## Mapping metrics types from v3

* GAUGE / RATE

  No changes here. Both types work in the same way as before.

* PRATE

  It maps to the new `cumulative-rate` metric.

* DELTA / PDELTA

  They map to the new `cumulative-count` metric.
  
* ATTRIBUTES

  This metric type has been removed. Now every metric can have attributes, called dimensions, attached to it.

#### Summary

| type           | v3 | v4 | mapping                                                                       |
|----------------|----|----|-------------------------------------------------------------------------------|
| gauge          | ✅ | ✅ | integration.Gauge(timestamp, name, value)                                     |
| rate           | ✅ | ✅ | integration.Rate(timestamp, name, value)                                      |
| prate          | ✅ | ❌ | integration.CumulativeRate(timestamp, name, value)                            |
| delta / pdelta | ✅ | ❌ | integration.CumulativeCount(timestamp, name, count)                           |
| attribute      | ✅ | ❌ | add dimensions to the metrics                                                 |
| count          | ❌ | ✅ | integration.Count(timestamp, interval, name, count)                           |
| summary        | ❌ | ✅ | integration.Summary(timestamp, interval, name, count, average, sum, min, max) |

#### Adding dimensions to a metric

Metrics can have a list of dimensions. These dimensions, along with the name, make the metric unique and can be used to 
search and filter them.

To add a new Dimensions just call on a metric 

`func (i *<Gauge|Count|Summary>) AddDimension(key string, value interface{})`

If the `key` already exists the previous value will be overwritten with the new one.

#### Adding metrics to an Entity

Once you have a metric, it can be attached to an Entity. The metric will belong to that Entity and it will be decorated
with the Entity's tags.

To add a metric to an entity just call the method on an Entity

`func (e *Entity) AddMetric(metric)`

### Events

The API for creating events has changed a little bit. Events can be attached to an Entity or Integration.

In SDK v3, events were created like this

`func New(summary, category string) *Event`

Now, for creating an event, you need to call this method inside the `integration` package

`func NewEvent(timestamp time.Time, summary, category string) *Event`

The parameters are the same as before with the exception of the `timestamp`. `summary` cannot be empty.
 
#### Adding attributes to an Event

In SDK v3, attributes were added to events using this method

`func AddCustomAttributes(e *Event, customAttributes []attribute.Attribute)`

In SDK v4, you need to use this other method for adding a single attribute

`func (e *Event) AddAttribute(key string, value interface{})`

If the `key` already exists the previous value will be overwritten with the new one.

#### Adding an event to an Entity

In both SDKs, the same method exists for adding an event to an entity

`func (e *Entity) AddEvent(event *integration.Event) error`

The event will be decorated with the Entity's tags.

### Inventory

The API for creating inventory is almost the same. Inventory can be attached to an Entity or Integration.

In SDK v3, an inventory item was added like this

`func (e *Entity) SetInventoryItem(key string, field string, value interface{}) error`

Now, to add an inventory item to an integration you use the following method

`func (e *Entity) AddInventoryItem(key string, field string, value interface{}) error`

Basically you just need to replace `SetInventoryItem` with `AddInventoryItem`. 


[1]: https://docs.newrelic.com/docs/data-ingest-apis/get-data-new-relic/metric-api/introduction-metric-api