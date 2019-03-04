# What is an entity?

We use the vague term entity because we want to support hosts, pods, load balancers, DBs, etc. in a generic way. In the previous SDK versions (v1 & v2) the entity was local and just one, the host. In later versions the host reporting is called local entity, and it's optional to add metrics to it. You could just use remote entities to attach metrics. An entity can have its own inventory (configuration/state) and report any kind of metrics about itself. Although we may define a new entity for each monitored thing, we may want to relate/group some of them within a parent (local) entity (ie: host it's running on).
 
Given we wanted to offer support for remote entities,let’s start by defining what we mean by an entity. **Entity** is a general term and refers to is a specific thing we collect data about. We used this term generally because we want to monitor different services and will need to support different parts such as hosts, pods, load balancers, DBs, etc. In the previous SDK versions (v1 & v2) the entity was local and just one, the host. However, as the entity does not necessarily have to be a host, we have broadened our support to include remote entities.
 
## What do we mean by remote entity?

We have defined an entity is defined as anything we collect data about. Previously, the entity was the local machine. In the new version, the **local entity** is the the local host. This leaves us to define what we mean by **remote entities**.
 
So what do we mean by remote entities? A remote entity is an entity, as understood above, that 1) is not the local host where either the agent and integration live or 2) is not equivalent to a “host”. In this way, we offer support to monitor a whole host of things, forgive the bad pun, that we might not otherwise be able to, due to being able to get an agent on the host or the architecture itself requiring the concept of multiple “entities”.
 
> For Example:
>
> An engineer installs the infrastructure agent on host1 and configures the out-of-the-box mysql integration to monitor a MySQL server running on host2. Host1 would be an entity as far as the Infrastructure entity is concerned and the MySQL servers on Host2 would be a “remote entity”.
>
> An engineer installs the infrastructure agent on host1 and configures the out-of-the-box kubernetes integration and collects metrics and inventory about the whole cluster, replica sets, pods and nodes.


## Entity uniqueness

Entity uniqueness is provided by its **key**.

An entity `key` is formed by its `namespace` (or *type*) and `name`. These fields are concatenated with `:` separator.

> For instance: `integration.Entity("name", "ns")` would create an entity with a `key` value `ns:name`.   

Entities are attached to a user account. Each integration is responsible for providing uniqueness for its entities.

It's up to the integration developer to define an entity `name` that's **unique** to its user account. An entity name should map uniquely to the entity being monitored. Imagine a situation where you are monitoring two Kafka clusters, one meant for development and the other for production, where you have a topic with the same name on each. The entity needs to map uniquely to the development topic and the production topic so that there is no collision.


## Entity naming guidelines

An entity name is an string identifier by which user uniquely identifies an entity.

Entity names should:

- provide enough information to be uniquely identified
- easily identifiable for a user to locate what they refer to

**Key point**: bundle as many identifier-attributes as required but not more.

Usually an *endpoint* is a good starting point to identify a service.

> Eg: `host:port`


### Entity name composition

Names could be composed of several fields, in this case we call these **identifier-attributes**.

#### Identifier attributes

Attributes that provide uniqueness to an entity.  

> Eg: If your endpoint gather several environments, then the endpoint string is not enough to identify a service if
> you want to attach different data to each environment.

For **identifier-attributes** add both the attribute *key* and *value* to the `name` to ease readability.

For **endpoints** just use the endpoint format (host:port) to start the `name` string.

A usual composition is: *endpoint + [identifier attribute key + identifier attribute value]*. With as many identifier
attributes as required.

> Eg: `endpoint:identifier_attr_key=identifier_attr_value`


### Name Format

Names are flexible, so random strings are accepted, but some good standards are:

- Composed fields are separated by colon `:`.
- Identifier-attributes have key & value, these are separated by equal sign `=`.

> Eg: `host:port:environment=prod`

We use this format to build the `name` when entities are built using the SDK `integration.Entity()` function.


### SDK entity creation

Function `integration.Entity(namespace, name, ...identifierAttrs)` ease all of this for you.

When you provide identifier-attributes:

- It automatically attach all the attributes to the entity `key` as defined above.
- It decorates all of the attached entity metrics with these attributes.
  * So these metrics could be easily filtered in the UI or New Relic Insights. 


## Metadata decoration

Since [agent v1.0.1052](https://docs.newrelic.com/docs/release-notes/infrastructure-release-notes/infrastructure-agent-release-notes/clone-new-relic-infrastructure-agent-101052) entity metrics could be decorated with metadata.

At the moment integrations can only decorate their entities metrics with `hostname`, `clusterName` and `serviceName`.
These are all optional decoration values provided via [args](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/args.md)


### Custom provided metadata

Any integration can be decorated by enabling `metadata` flag on [args](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/args.md) package.

If this flag is enabled all samples will be decorated with provided key-value pairs.

At the moment metadata could only be provided via environment variables. These should be prefixed with `NRI_` and value
will be provided after an equal sign `=` separator. 

Ie: `NRI_{INTEGRATION_NAME}_FOO=BAR` will decorate samples with a `FOO` attribute holding a `BAR`value.
