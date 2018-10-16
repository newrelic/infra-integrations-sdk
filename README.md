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
$ go get github.com/newrelic/infra-integrations-sdk
```

Then you can import any of the packages provided with the SDK from your code and
start writing your integration. If you need ideas or inspiration, you can follow [the tutorial](docs/tutorial.md).

## JMX support

The Integrations Go lang SDK supports getting metrics through JMX by calling the
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
[these guidelines](https://gopkg.in/newrelic/infra-integrations-sdk.v2/blob/master/CONTRIBUTING.md).

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
