# New Relic Infrastructure Integrations SDK

New Relic Infrastructure, provided by New Relic, Inc (http://www.newrelic.com),
offers flexible, dynamic server monitoring. We provide an SDK for creating an
integration for reporting custom host and server data, including metric, event,
and inventory (system state) data. That data will be findable and usable in New Relic
Infrastructure and in New Relic Insights. You can find the complete documentation
of the SDK on [our docs site](https://docs.newrelic.com/docs/intro-infrastructure-integration-sdk).

 The New Relic Integrations SDK is hosted on [github](https://github.com/newrelic/infra-integrations-sdk), and a
 the set of official integrations supported by New Relic are hosted on [github](https://github.com/newrelic/infra-integrations).

## Compatibility and requirements

Up-to-date [our docs site](https://docs.newrelic.com/docs/compatibility-requirements-infrastructure-integration-sdk).


## Contributing Code

We welcome code contributions (in the form of pull requests) from our user
community.  Before submitting a pull request please review
[these guidelines](https://github.com/newrelic/infra-integrations-sdk/blob/master/CONTRIBUTING.md).

Following these helps us efficiently review and incorporate your contribution
and avoid breaking your code with future changes to the agent.


## API specification

You can find the latest API documentation generated from the source code in
[godoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk).

## Getting Started

Before starting to write Go code, we suggest taking a look at
[golang's documentation](https://golang.org/doc/code.html) to setup the
environment and familiarize yourself with the golang language.

You can download the SDK code to your GOPATH with the following command:

```bash
$ go get github.com/newrelic/infra-integrations-sdk
```

Then you can import any of the packages provided with the SDK from your code and
start writing your integration. You can find some existing integrations provided
by New Relic in our [github](https://github.com/newrelic/infra-integrations)

## Support

You can find more detailed documentation [on our website](http://newrelic.com/docs),
and specifically in the [Infrastructure category](https://docs.newrelic.com/docs/infrastructure).

If you can't find what you're looking for there, reach out to us on our [support
site](http://support.newrelic.com/) or our [community forum](http://forum.newrelic.com)
and we'll be happy to help you.

Find a bug? Contact us via [support.newrelic.com](http://support.newrelic.com/),
or email support@newrelic.com.

New Relic, Inc.
