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
