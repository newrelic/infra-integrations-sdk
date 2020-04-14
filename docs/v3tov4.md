# Upgrading from GoSDK v3.x to GoSDK v4.x

V4 of the GoSDK is not backward compatible with the previous versions of the GoSDK. This document describes the changes 
between versions and how to adapt an integration built with the GoSDK v3.x to the new one.

The Go SDK v4 contains the following changes:

* New Infrastructure Agent Integration JSON version 4.
* Add support for dimensional metrics using the [Metrics API format][1].
* New metric data types: `count` and `summary`.
* Removed support for `RATE`, `PRATE` and `DELTA` metric types.
* Removed distinction between remote and local entities.
* Metrics, Events and Inventory can be attached either to Entities or an Integration.
* Support for Go Modules
* Removed support for protocols prior to v4.x.

## JSON schema changes

The following section explains the new JSON schema. Be aware that the new SDK only supports this new JSON document. 
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
      "entity":{                              # this object is optional
        "name":"redis:192.168.100.200:1234",  # unique entity name per account
        "type":"RedisInstance",               # entity's category
        "displayName":"my redis instance",    # human readable name
        "tags":{}                             # key-value pairs that will be also added as attributes to all metrics 
                                              # and events
      },
      "metrics":[                             # list of metrics using the dimensional metric format
        {
          "name":"redis.metric1",
          "type":"count",                     # gauge, count, summary or pdelta
          "value":93, 
          "attributes":{}                     # set of key-value pairs that define the dimensions of the metric
        }
      ],
      "inventory":{...},                      # Inventory data format has not changed
      "events":[...]                          # Events data format has not changed
    }
  ]
}
```

## GoSDK v4 API changes

This section enumerates the main changes you have to keep in mind if you want to upgrade from GoSDK v3.x to v4.x

### Entities

In previous versions of the SDK, we had a distinction between Local and Remote entities. You could get the 
`Local Entity` calling the method

`func (i *Integration) LocalEntity() *Entity`

or create a `Remote Entity` calling

`func (i *Integration) Entity(name, namespace string, idAttributes ...IDAttribute) (e *Entity, err error)`

In SDK v4, there's only one way of creating entities:

`func (i *Integration) Entity(name, type, displayName string) (e *Entity, err error)`

* `name` must be an unique value per customer account because it uniquely identifies your Entity. It cannot be empty and 
  the SDK cannot validate nor enforce that uniqueness, so it's up to the client to define a naming schema that produces 
  unique names.

  For example, using the following schema `<service_name>:<ip|hostname>:<port>`, we could create these names: 
  `redis:192.168.100.200:12345` or `mysql:my_server_hostname:8086`

  For the ip or the hostname, please avoid using the values `127.0.0.1` or `localhost` as they don't produce unique 
  identifiers.

* `type` describes the Entity's category. It cannot be empty. Some examples: "RedisInstance", "DockerContainer", 
  "K8SPod", "CassandraCluster", "MySQLInstance", "KafkaBroker", "NginxServer", etc.

* `displayName` is a friendly human readable name to be used in the UI. It can be empty.

#### Migrating from LocalEntity

If you are using the LocalEntity, you need to use the new Entity method with the appropriate parameters.

#### Adding tags to an Entity

Entities can have a list of tags. These tags can be used to search and filter your entities and will be added to 
all the metrics and events of the Entity.

To add a new Tag just call 

`func (i *Entity) AddTag(key string, value interface{})`

If the `key` already exists, then the previous value will be overwritten with the new one.

### Metrics

#### Types

In this SDK v4 we introduce some new metric types while removing some others.

* GAUGE

  No changes here. This metric type works in the same manner. 

* RATE and PRATE

  Both metric types have been removed. You should use a `gauge` value and replicate the behaviour using the NRQL `rate` 
  function:

  `FROM Metric SELECT rate(average(metricName), 1 minute)`

* DELTA

  The `delta` metric type has been removed. You should use a `gauge` value instead and take care of the delta 
  calculation.

* PDELTA

  This metric type still exists and works in the same manner.
  
* ATTRIBUTES

  This metric type has been removed. Now every metric can have attributes, called dimensions, attached to it.

* COUNT

  This metric type is new. Measures the number of occurrences of an event given a time interval.

  Examples include cache hits per reporting interval and the number of threads created per reporting interval.

* SUMMARY

  This metric type is new. Used to report pre-aggregated data, or information on aggregated discrete events. A summary 
  includes a count, sum value, min value, and max value. 

  Examples include transaction count/durations and queue count/ durations.

#### Creating v4 metrics

| metric  | method                                                                        |
|---------|-------------------------------------------------------------------------------|
| Gauge   | integration.Gauge(timestamp, name, value)                                     |
| Count   | integration.Count(timestamp, interval, name, count)                           |
| Summary | integration.Summary(timestamp, interval, name, count, average, sum, min, max) |
| PDelta  | integration.PDelta(timestamp, name, value)                                    |

##### Definition of the parameters

`timestamp` is the metric's start time in Unix time (milliseconds).

`name` is the name of the metric. The value must be less than 255 characters.

`value` is a double. Gauge values can be positive or negative. PDelta only allows positive values.

`interval` is the length of the time window in milliseconds.

`count` is the number of occurrences of an event reported in a given interval. Must be a positive double.

`average` is the number expressing the central value in the set of data.

`sum` is the aggregation of all the values registered in the set of data.

`min` is the minimum value registered in the set of data.

`max` is the maximum value registered in the set of data.

#### Adding dimensions to a metric

Metrics can have a list of dimensions. These dimensions, along with the name, make the metric unique and can be used to 
search and filter them.

To add a new Dimensions just call on a metric 

`func (i *<Gauge|Count|Summary>) AddDimension(key string, value interface{})`

If the `key` already exists, then the previous value will be overwritten with the new one.

#### Adding metrics to an Entity

Once you have a metric, it can be attached to an Entity. The metric will belong to that Entity and it will be decorated
with the Entity's tags.

To add a metric to an entity just call the method on an Entity

`func (e *Entity) AddMetric(metric)`

#### Adding metrics to an Integration

Metrics can also be added to an Integration. Those metrics won't be attached to any Entity.

To add a metric to an Integrations just call the method inside the `integration` namespace

`func (e *Integration) AddMetric(metric)`

#### Metrics summary

| type       | v3 | v4 | mapping                                                                          |
|------------|----|----|---------------------------------------------------------------------------------|
| gauge      | ✅  | ✅  | integration.Gauge(timestamp, name, value)                                     |
| rate/prate | ✅  | ❌  | use a gauge and the NRQL rate function                                        |
| delta      | ✅  | ❌  | use a gauge and calculate the delta                                           |
| pdelta     | ✅  | ✅  | integration.PDelta(timestamp, name, value)                                    |
| attribute  | ✅  | ❌  | add dimensions to the metrics                                                 |
| count      | ❌  | ✅  | integration.Count(timestamp, interval, name, count)                           |
| summary    | ❌  | ✅  | integration.Summary(timestamp, interval, name, count, average, sum, min, max) |

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

If the `key` already exists, then the previous value will be overwritten with the new one.

#### Adding an event to an Entity

In both SDKs, the same method exists for adding an event to an entity

`func (e *Entity) AddEvent(event *integration.Event) error`

The event will be decorated with the Entity's tags.

#### Adding an event to an Integration

An Event can also be attached to an Integration which means that the event doesn't belong to any specific entity. You
can call this method

`func (i *Integration) AddEvent(event *integration.Event) error`

### Inventory

The API for creating inventory is almost the same. Inventory can be attached to an Entity or Integration.

In SDK v3, an inventory item was added like this

`func (e *Entity) SetInventoryItem(key string, field string, value interface{}) error`

Now, to add an inventory item to an integration you use the following method

`func (e *Entity) AddInventoryItem(key string, field string, value interface{}) error`

Basically you just need to replace `SetInventoryItem` with `AddInventoryItem`. 

#### Adding inventory to an Integration

Inventory can also be added to an Integration. The previous call can be used on an Integration 

`func (e *Integration) AddInventoryItem(key string, field string, value interface{}) error`


[1]: https://docs.newrelic.com/docs/data-ingest-apis/get-data-new-relic/metric-api/introduction-metric-api