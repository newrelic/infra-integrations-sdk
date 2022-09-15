[![New Relic Community Plus header](https://raw.githubusercontent.com/newrelic/open-source-office/master/examples/categories/images/Community_Plus.png)](https://opensource.newrelic.com/oss-category/#community-plus)

[![BuildStatus Widget]][BuildStatus Result]
[![GoReport Widget]][GoReport Status]
[![GoDocWidget]][GoDocReference]

[BuildStatus Result]: https://travis-ci.org/newrelic/infra-integrations-sdk
[BuildStatus Widget]: https://travis-ci.org/newrelic/infra-integrations-sdk.svg?branch=master

[GoReport Status]: https://goreportcard.com/report/github.com/newrelic/infra-integrations-sdk
[GoReport Widget]: https://goreportcard.com/badge/github.com/newrelic/infra-integrations-sdk

[GoDocReference]: https://godoc.org/github.com/newrelic/infra-integrations-sdk
[GoDocWidget]: https://godoc.org/github.com/newrelic/infra-integrations-sdk?status.svg

# Golang SDK for New Relic integrations

[Infrastructure monitoring](https://docs.newrelic.com/docs/infrastructure) provided by [New Relic](http://www.newrelic.com) offers flexible, dynamic server monitoring, including [integrations](https://docs.newrelic.com/docs/integrations/new-relic-integrations/get-started/introduction-infrastructure-integrations) for many popular services.

If our [on-host integrations](https://docs.newrelic.com/docs/integrations/host-integrations/getting-started/introduction-host-integrations) don't meet your needs, we provide two options for creating your own:

* Our [Flex integration tool](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/flex-integration-tool-build-your-own-integration): a simple way to report custom metrics by creating a configuration file that defines what data will be reported. This is recommended for most use cases.
* Our Integrations SDK: a more robust solution. We give you access to the complete set of tools we use to build our integrations and report all [infrastructure integrations data types](https://docs.newrelic.com/docs/integrations/new-relic-integrations/get-started/introduction-infrastructure-integrations#data-types).

The Integrations SDK helps take the complexity out of building an integration by providing a set of useful Go language functions and data structures. For instance, some common use cases like reading values from command-line arguments or environment variables, initializing a structure with all the necessary fields for an integration defined by our SDK, or generating and printing a JSON to stdout, are covered and simplified by this package.

If you want to know more or you need specific documentation about the structures and functions provided by this package, you can take a look at the official package documentation in godoc.org (see below).

# SDK v4 Internal Release Notice

This is an internal release of the new SDK v4. It contains breaking changes, therefore it's highly recommended to take
a look at the [migration guide from v3 to v4](#upgrading-from-SDK-v3-to-v4).

Most of the documentation hasn't been updated yet to reflect the changes made in this new release.

## Installation

Before starting to write Go code, we suggest taking a look at [golang's documentation](https://golang.org/doc/code.html) to set up the environment and familiarize yourself with the golang language.

The minimum supported Go version is 1.13. You can check your Go version executing the following command in a bash shell:

```bash
$ go version
```

You can download the SDK code to your `GOPATH` with the following command:

```bash
$ go get github.com/newrelic/infra-integrations-sdk/...
```

Read the [SDK documentation](docs/README.md) to learn about all the packages and functions it provides. If you need ideas or inspiration to start writing integrations, follow [the tutorial](docs/tutorial.md).

## API specification

You can find the latest API documentation generated from the source code in [godoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk).

### Agent API

Infrastructure on-host integrations are executed periodically by the infrastructure agent. The integration `stdout` is consumed by the agent. `stdout` data is formatted as JSON.

The agent supports different JSON data-structures called *integration protocols*:

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

https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/v3tov4.md

### SDK and agent-protocol compatibility

SDK v1 and v2 use *protocol-v1*.

SDK v3 could use either *protocol-v2* or *protocol-v3*.

SDK v4 only uses *protocol-v4*.

## Libraries

### JMX support

The Integrations Golang SDK supports getting metrics through JMX by calling the `jmx.Open()`, `jmx.OpenWithSSL`, `jmx.Query()`, and `jmx.Close()` functions. This JMX support relies
on the nrjmx tool. Follow the steps in the [nrjmx](https://github.com/newrelic/nrjmx) repository to build it and set the `NR_JMX_TOOL` environment variable to point to the location of the nrjmx
executable. If the `NR_JMX_TOOL` variable is not set, the SDK uses `/usr/bin/nrjmx` by default.

### HTTP support

The GoSDK provides a helper HTTP package to create secure HTTPS clients that require loading credentials from a Certificate Authority Bundle (stored in a file or in a directory). You can read more [here](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/http.md).

## Tools and FAQs related to previous SDK v3.

https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/v3_tools_faqs.md

## Support

Should you need assistance with New Relic products, you are in good hands with several support diagnostic tools and support channels.

If the issue has been confirmed as a bug or is a feature request, file a GitHub issue.

**Support Channels**

* [New Relic Documentation](https://docs.newrelic.com): Comprehensive guidance for using our platform
* [New Relic Community](https://discuss.newrelic.com): The best place to engage in troubleshooting questions
* [New Relic Developer](https://developer.newrelic.com/): Resources for building a custom observability applications
* [New Relic University](https://learn.newrelic.com/): A range of online training for New Relic users of every level
* [New Relic Technical Support](https://support.newrelic.com/) 24/7/365 ticketed support. Read more about our [Technical Support Offerings](https://docs.newrelic.com/docs/licenses/license-information/general-usage-licenses/support-plan).

## Privacy

At New Relic we take your privacy and the security of your information seriously, and are committed to protecting your information. We must emphasize the importance of not sharing personal data in public forums, and ask all users to scrub logs and diagnostic information for sensitive information, whether personal, proprietary, or otherwise.

We define “Personal Data” as any information relating to an identified or identifiable individual, including, for example, your name, phone number, post code or zip code, Device ID, IP address, and email address.

For more information, review [New Relic’s General Data Privacy Notice](https://newrelic.com/termsandconditions/privacy).

## Contribute

We encourage your contributions to improve this project! Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

If you have any questions, or to execute our corporate CLA (which is required if your contribution is on behalf of a company), drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project would not be what it is today.

## License

infra-integrations-sdk is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
