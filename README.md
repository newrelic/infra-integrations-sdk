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

# SDK v4 Internal Release Notice

This is an internal release of the new SDK v4. It contains breaking changes, therefore it's highly recommended to take
a look at the [migration guide from v3 to v4](#upgrading-from-SDK-v3-to-v4).

Most of the documentation hasn't been updated yet to reflect the changes made in this new release.

## Getting Started

Before starting to write Go code, we suggest taking a look at
[golang's documentation](https://golang.org/doc/code.html) to setup the
environment and familiarize yourself with the golang language. 

The minimum supported Go version is 1.13. You can check your Go version executing the following command in a bash shell:

```bash
$ go version
```

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
* v4: Adds support for dimensional metrics format and introduces new metric types: `count`, `summary`, `cumulative-count`
 and `cumulative-rate`.

### Host Entity vs Entities

`Entity` is a specific thing we collect data about. We used this vague term because we want to support hosts, pods, load
 balancers, DBs, etc. in a generic way. In the previous SDK v3, we had the Local Entity and Remote Entities. 
 
In this new version the reporting host is called **HostEntity**, and it's optional to add data to it. It represents the 
host where the agent is running on. If your entity belongs to a different host or it's something abstract that is 
not attached to the host where the integration runs, then you can create an Entity which requires a unique name and 
an entity type in order to be created.

You can add metrics, events and inventory on both types of entities.

## Upgrading from SDK v3 to v4

https://github.com/newrelic/infra-integrations-sdk/blob/sdk-v4/docs/v3tov4.md

#### SDK & agent-protocol compatibility 

SDK v1 and v2 use *protocol-v1*.

SDK v3 could use either *protocol-v2* or *protocol-v3*.

SDK v4 only uses *protocol-v4*.

## Libraries

### JMX support

The Integrations Go lang SDK supports getting metrics through JMX by calling the `jmx.Open()`, `jmx.OpenWithSSL`, 
`jmx.Query()` and `jmx.Close()` functions. This JMX support relies on the nrjmx tool. Follow the steps in the [nrjmx](https://github.com/newrelic/nrjmx) 
repository to build it and set the `NR_JMX_TOOL` environment variable to point to the location of the nrjmx executable. 
If the `NR_JMX_TOOL` variable is not set, the SDK will use `/usr/bin/nrjmx` by default.

### HTTP support

The GoSDK provides a helper HTTP package to create secure HTTPS clients that require loading credentials from a 
Certificate Authority Bundle (stored in a file or in a directory). You can read more [here](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/http.md).

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

## Tools and FAQs related to previous SDK v3.

https://github.com/newrelic/infra-integrations-sdk/blob/sdk-v4/docs/v3_tools_faqs.md
