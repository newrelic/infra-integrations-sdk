# New Relic Infrastructure Integrations - Go lang SDK

[![BuildStatus Widget]][BuildStatus Result]
[![GoReport Widget]][GoReport Status]
[![GoDocWidget]][GoDocReference]

[BuildStatus Result]: https://travis-ci.org/newrelic/infra-integrations-sdk
[BuildStatus Widget]: https://travis-ci.org/newrelic/infra-integrations-sdk.svg?branch=master

[GoReport Status]: https://goreportcard.com/report/github.com/newrelic/infra-integrations-sdk
[GoReport Widget]: https://goreportcard.com/badge/github.com/newrelic/infra-integrations-sdk

[GoDocReference]: https://godoc.org/github.com/newrelic/infra-integrations-sdk
[GoDocWidget]: https://godoc.org/github.com/newrelic/infra-integrations-sdk?status.svg


New Relic Infrastructure, provided by New Relic, Inc (http://www.newrelic.com),
offers flexible, dynamic server monitoring. We provide an SDK for creating an
integration for reporting custom host and server data, including metric, event,
and inventory (system state) data. That data will be findable and usable in New Relic
Infrastructure and in New Relic Insights. You can find the complete documentation
of the SDK on [our docs site](https://docs.newrelic.com/docs/intro-infrastructure-integration-sdk).

This Go package helps take the complexity out of building an Infrastructure
Integration by providing a set of useful GO functions and data structures. For
instance, some common use cases like reading values from command-line arguments
or environment variables, initializing a structure with all the necessary fields
for an Integration defined by our SDK or generating and printing a JSON to
stdout, are covered and simplified by this package.

If you want to know more or you need specific documentation about the structures
and functions provided by this package, you can take a look at the official
package documentation in godoc.org (see below).

## Getting Started

Before starting to write Go code, we suggest taking a look at
[golang's documentation](https://golang.org/doc/code.html) to setup the
environment and familiarize yourself with the golang language.

You can download the SDK code to your GOPATH with the following command:

```bash
$ go get github.com/newrelic/infra-integrations-sdk/...
```

And then you can read the [SDK comprehensive documentation](docs/README.md) to know all the packages and functions
it provides to start writing your integration. If you need ideas or inspiration, you can follow
[the tutorial](docs/tutorial.md).

## API specification

You can find the latest API documentation generated from the source code in
[godoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk).

### Agent API

Integrations are executed periodically by the *agent*. The integration `stdout` is consumed by the *agent*.
This `stdout` data is formatted as JSON.

Agent supports different JSON data-structures called *integration protocols*:

* v1: Legacy data structure to monitor local entity.
* v2: This version allows to monitor remote entities and keep support for previous local entity. [Official doc](https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/integration-executable-file-specifications)
* v3: Improves remote entities support. See [protocol v3](docs/protocol-v3.md) documentation. 


### Local Entity vs Remote Entities

`Entity` is a specific thing we collect data about. We used this vague term because we want to support hosts, pods, load balancers, DBs, etc. in a generic way. In the previous SDK versions (v1 & v2) the entity was local and just one, the host.

In later versions the host reporting is called **local entity**, and it's optional to add metrics to it. You could just use **remote entities** to attach metrics.

For more information on the definition of a remote entity, please see the following document on [local vs remote entites](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/entity-definition.md).



## Upgrading from SDK v2 to v3

https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/v2tov3.md

#### SDK & agent-protocol compatibility 

SDK v1 and v2 use *protocol-v1*.

SDK v3 could use either *protocol-v2* or *protocol-v3*.


## Tools

This section shows the documentation of all the core components of the GoSDK v3. This is, all the packages that are required to setup and manage the integration data, as well as some other libraries that, despite they are not part of the core SDK, implement common operations that may be reused in different integrations:

### Command-line tool to scaffold new New Relic custom integrations

https://github.com/newrelic/nr-integrations-builder

### Integration data model

https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/integration.md


### Configuration arguments

https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/args.md

### Internal logging

https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/log.md

### Persistence/Key-Value storage

https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/persist.md


### Legacy protocol v1 builder

In case you want to use the previous builder, you still can do it via `gopkg.in/newrelic/nr-integrations-builder.v1`.

You can easily fetch it using:
 
`go get gopkg.in/newrelic/nr-integrations-builder.v1`

## Libraries

### JMX support

The Integrations Go lang SDK supports getting metrics through JMX by calling the
`jmx.Open()`, `jmx.OpenWithSSL`, `jmx.Query()` and `jmx.Close()` functions. This JMX support relies
on the nrjmx tool. Follow the steps in
the [nrjmx](https://github.com/newrelic/nrjmx) repository to build it and set
the `NR_JMX_TOOL` environment variable to point to the location of the nrjmx
executable. If the `NR_JMX_TOOL` variable is not set, the SDK will use
`/usr/bin/nrjmx` by default.

### HTTP support

GoSDK v3 provides a helper HTTP package to create secure HTTPS clients that require loading credentials from a Certificate Authority Bundle (stored in a file or in a directory).You can read more [here](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/http.md).


## Contributing Code

We welcome code contributions (in the form of pull requests) from our user
community.  Before submitting a pull request please review
[these guidelines](https://github.com/newrelic/infra-integrations-sdk/blob/master/CONTRIBUTING.md).

Following these helps us efficiently review and incorporate your contribution
and avoid breaking your code with future changes to the agent.

## Support

You can find more detailed documentation [on our website](http://newrelic.com/docs),
and specifically in the [Infrastructure category](https://docs.newrelic.com/docs/infrastructure).

If you can't find what you're looking for there, reach out to us on our [support
site](http://support.newrelic.com/) or our [community forum](http://forum.newrelic.com)
and we'll be happy to help you.

Find a bug? Contact us via [support.newrelic.com](http://support.newrelic.com/),
or email support@newrelic.com.

New Relic, Inc.


## FAQs

- **Where can I see the data I'm sending to New Relic with my custom integration?**

    * You can go to [insights](https://insights.newrelic.com/) and run [these queries]( https://github.com/newrelic/infra-integrations-sdk/blob/faqs/docs/tutorial.md#view-metric-data-in-new-relic-insights) to see metric data.
    * In order to see inventory data follow [these intructions](https://github.com/newrelic/infra-integrations-sdk/blob/faqs/docs/tutorial.md#view-inventory-data-in-infrastructure).
    
    
- **Are there other SDKs in other languages?**

    * No, we plan to have more in the future though.
- **Do you have examples of integrations written in other languages?**

    * Yes, you can find them at this [link](https://github.com/newrelic/infra-integrations/tree/master/doc/examples).

- **Can I reuse my Nagios checks?**

    * We don't offer an Integration for "Nagios" checks. But with the SDK you can build a simple custom integration to push your check data

- **How can I migrate my current integration from sdk v2 to sdk v3?**
    
    * You can check our [migration steps](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/v2tov3.md).

<!--
 - **Can I attach a custom calculation function to a metric type?** 
-->

- **How can I add a custom argument to my integration?**

    * Take a look at [arguments documentation](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/args.md).

- **How can I query a server that uses a custom certificate?**

    * We have [http helper](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/http.md) that you can use to setup custom certificates easily.

- **How can I query JMX bean using the sdk?**

    * We have a [Jmx helper](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/jmx.md) that will help you query any beam you are interested in monitoring.

        



