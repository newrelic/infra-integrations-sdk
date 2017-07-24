# New Relic Infrastructure Integrations - Go lang SDK

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

Besides the idea of provide this library for helping people to write their
custom integrations, New Relic is using it for writing the official set
of
[Infrastructure Integrations](https://github.com/newrelic/infra-integrations).
For this reason, the package will evolve continuously and we'll improve or add
new characteristics, trying to simplify even further the process of writing an
Integration.


## Getting Started

Before starting to write Go code, we suggest taking a look at
[golang's documentation](https://golang.org/doc/code.html) to setup the
environment and familiarize yourself with the golang language.

You can download the SDK code to your GOPATH with the following command:

```bash
$ go get github.com/newrelic/infra-integrations-sdk
```

Then you can import any of the packages provided with the SDK from your code and
start writing your integration. If you need ideas or inspiration, you can follow [the tutorial](docs/tutorial.md) or find
some existing integrations provided by New Relic in
our [github](https://github.com/newrelic/infra-integrations)

## JMX support

The Integrations Go lang SDK supports getting metrics through JMX through the
`jmx.Open()`, `jmx.Query()` and `jmx.Close()` functions. This JMX support relies
on the nrjmx tool. Follow the steps in
the [nrjmx](https://github.com/newrelic/nrjmx) repository to build it and set
the `NR_JMX_TOOL` environment variable to point to the location of the nrjmx
executable. If the `NR_JMX_TOOL` variable is not set, the SDK will use
`/usr/bin/nrjmx` by default.

## API specification

You can find the latest API documentation generated from the source code in
[godoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk).

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
