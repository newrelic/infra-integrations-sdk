# JMX package

`GoSDK v3` allows building integrations that query information from the
[Java Management eXtensions](http://www.oracle.com/technetwork/articles/java/javamanagement-140525.html) by means of
the `jmx` package (which requires to install the [NR JMX tool](https://github.com/newrelic/nrjmx)).

## Installing NR-JMX

### Linux

Assuming you have already installed the
[New Relic Infrastructure Agent](https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation),
or at least you have added the required package repositories as it is described in the linked documentation, you have
to install the `nrjmx` package, this is:

`sudo yum install nrjmx` (for RedHat-based systems)

`sudo apt-get install nrjmx` (for Debian-based systems)

(SUSE-based systems will be supported in the future)

### Windows

At this moment, NR-JMX is not supported for Windows, but in the near future we will release an installer executable.

## Using the `jmx` client from your integration

This section explain the basic usage of the JMX helper functions. For a more detailed description, visit
the [SDK JMX GoDoc page]([http.New GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/jmx)).

* `jmx.Open` ([GoDoc]([http.New GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/jmx#Open)))
    - Instantiates a JMX client. It requires a `hostname` and a `port` as arguments. `username` and
      `password` are optional (pass empty strings `""` if no user/password).
    - By default, the generated client will look for the [NRJMX tool](#installing-nr-jmx) in the path `/usr/bin/nrjmx`,
      but this location can be overriden by means of the `NR_JMX_TOOL` environment variable.
    - This function will return the global `ErrJmxCmdRunning` error, if there is already another instance of NRJMX
      running.
* `jmx.Close` ([GoDoc]([http.New GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/jmx#Close)))
    - This function must be invoked to close the JMX connection and free the associated resources.
* `jmx.Query` ([GoDoc]([http.New GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/jmx#Query)))
    -  Submits a JMX query as an object pattern string, returning the results as a map, where the `key` represents the
       MBean Object Name (in the `domain:key-property-list` form) and the value is the sample value for this metric
       at the given moment.
       
## Basic usage example

```go
jmx.Open("127.0.0.1", "9010", "", "")

results, _ := jmx.Query("java.lang:type=OperatingSystem", 5 * time.Second) // 5s timeout

for key, value := range results {
    fmt.Printf("%v -> %v", key, value)
}

jmx.Close() 
```

The above code would print something similar to:
```
java.lang:type=OperatingSystem,attr=SystemLoadAverage -> 1.59423828125
java.lang:type=OperatingSystem,attr=Arch -> "x86_64"
java.lang:type=OperatingSystem,attr=OpenFileDescriptorCount -> 35
java.lang:type=OperatingSystem,attr=ProcessCpuLoad -> 1.873180382851263E-4
java.lang:type=OperatingSystem,attr=MaxFileDescriptorCount -> 10240
java.lang:type=OperatingSystem,attr=CommittedVirtualMemorySize -> 10359709696
java.lang:type=OperatingSystem,attr=FreePhysicalMemorySize -> 1213214720
java.lang:type=OperatingSystem,attr=TotalSwapSpaceSize -> 1073741824
java.lang:type=OperatingSystem,attr=Name -> "Mac OS X"
java.lang:type=OperatingSystem,attr=Version -> "10.13.4"
java.lang:type=OperatingSystem,attr=TotalPhysicalMemorySize -> 17179869184
java.lang:type=OperatingSystem,attr=SystemCpuLoad -> 0.07889343680634009
java.lang:type=OperatingSystem,attr=AvailableProcessors -> 8
java.lang:type=OperatingSystem,attr=FreeSwapSpaceSize -> 977272832
java.lang:type=OperatingSystem,attr=ProcessCpuTime -> 1039197000
```
