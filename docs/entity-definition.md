# What is an entity?

We use the vague term entity because we want to support hosts, pods, load balancers, DBs, etc. in a generic way.

An entity can have its own inventory (configuration/state) and report any kind of metrics about itself.

**Entity** is a general term and refers to is a specific thing we collect data about. We used this term generally
because we want to monitor different services and will need to support different parts such as hosts, pods, load
balancers, DBs, etc.


## Local vs Remote entities

In the previous SDK versions (v1 & v2) the entity was local and just one, the host. In later versions the host
reporting/forwarding data is called **local entity**, and it represent the host running the agent.

Since SDK version v3 you can use a new entity for each monitored thing. These non-local entities are named
**remote entities**.

A **remote entity** is something being monitored, but it is no longer attached to the host running the agent. It
groups a set of metrics or storing inventory.

It's optional to add metrics to either the **local entity** or any **remote entities**.

For example:

> An engineer installs the infrastructure agent on *host1* and configures the out-of-the-box *mysql integration* to
> monitor a *MySQL server* running on *host2*. *Host1* would be a "local entity" and the *MySQL server* on *Host2*
> would be a “remote entity”.


### Host disclaimer

Historically New Relic infrastructure product has treated hosts as the only possible entities (see prior SDK versions
mentioned above).

For this reason some UI components may display the text *Hosts* when referring to *Entities*.


### Data sources

Optionally we might want to know/store the sources of an entity data.

We may want to differentiate:

- The entity **producing** a remote entity data
- The entity **reporting** a remote entity data

This gets relevance whenever the endpoint/entity reporting a remote-entity data is not the one producing it.

Usually for clusterised services, for instance being "cluster" and "nodes" entities. Then to retrieve a cluster-entity
an **endpoint** or a node (**entity**) report data about this cluster-entity.

> This might also apply to nodes reporting other nodes statues.

In this case we may want to know which entity or endpoint is reporting a given entity data.

The standard way to do it is using this *optional* attributes:

- `reportingEntityKey`: entity reporting actual entity data.
- `reportingEndpoint`: endpoint reporting current entity data.


## Entity uniqueness

Entity uniqueness is provided by its **key**.

An entity `key` is formed by its `namespace` (or *type*) and `name`. These fields are concatenated with `:` separator.

> For instance: `integration.Entity("name", "ns")` would create an entity with a `key` value `ns:name`.

Entities are attached to a user account. Each integration is responsible for providing uniqueness for its entities.

It's up to the integration developer to define an entity `name` that's **unique** to its user account. An entity name should map uniquely to the entity being monitored.

Imagine a situation where you are monitoring two Kafka clusters, one meant for development and the other for production, where you have a topic with the same name on each. The entity needs to map uniquely to the development topic and the production topic so that there is no collision.


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

In order to respect this format an identifier attribute should not use colon `:` or equals `=` characters in their name/values.

We use this format to build the `name` when entities are built using the SDK `integration.Entity()` function.


### SDK entity creation

Function `integration.Entity(namespace, name, ...identifierAttrs)` ease all of this for you.

When you provide identifier-attributes:

- It automatically attach all the attributes to the entity `key` as defined above.
- It decorates all of the attached entity metrics with these attributes.
  * So these metrics could be easily filtered in the UI or New Relic Insights.


## Metadata decoration

Since [agent v1.0.1052](https://docs.newrelic.com/docs/release-notes/infrastructure-release-notes/infrastructure-agent-release-notes/clone-new-relic-infrastructure-agent-101052)
entity metrics and events can be decorated with metadata.

Decoration by `hostname`, `clusterName` or `serviceName` can be set via
[args](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/args.md).

To decorate with custom provided metadata, the `metadata` flag should be enable
via [args](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/args.md).
If this flag is enabled all samples will be decorated with provided key-value
pairs.

At the moment custom metadata can only be provided via environment variables.
These should be prefixed with `NRI_{INTEGRATION_NAME}_`, were
`{INTEGRATION_NAME}` is the name given to the
[`Integration`](https://github.com/newrelic/infra-integrations-sdk/blob/master/integration/integration.go#L39).

Ie: `NRI_{INTEGRATION_NAME}_FOO=BAR` will decorate samples with a `FOO`
attribute holding a `BAR` value.
