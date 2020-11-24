# FAQ

- **Where can I see the data I'm sending to New Relic with my custom integration?**

    * Go to the [query builder](https://docs.newrelic.com/docs/query-your-data/explore-query-data/query-builder/introduction-query-builder) and run [these queries]( https://github.com/newrelic/infra-integrations-sdk/blob/faqs/docs/tutorial.md#view-metric-data-in-new-relic-insights).
    * To see inventory data, follow [these intructions](https://github.com/newrelic/infra-integrations-sdk/blob/faqs/docs/tutorial.md#view-inventory-data-in-infrastructure).
    
- **Are there other SDKs for other languages?**

    * No. We plan to have more in the future though.

- **Do you have examples of integrations written in other languages?**

    * Yes, you can find them [here](https://github.com/newrelic/infra-integrations/tree/master/doc/examples).

- **Can I reuse my Nagios checks?**

    * We don't offer an integration for Nagios checks. With the SDK you can build a simple custom integration to push your check data.

- **How can I migrate my current integration from SDK v2 to SDK v3?**
    
    * Check out our [migration steps](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/v2tov3.md).

- **How can I add a custom argument to my integration?**

    * Take a look at [arguments documentation](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/args.md).

- **How can I query a server that uses a custom certificate?**

    * We have an [http helper](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/http.md) that you can use to setup custom certificates easily.

- **How can I query JMX bean using the SDK?**

    * We have a [Jmx helper](https://github.com/newrelic/infra-integrations-sdk/blob/master/docs/toolset/jmx.md) that will help you query any beam you are interested in monitoring.