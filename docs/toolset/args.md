# Configuration arguments

The integrations `GoSDK v3` provides a tool to accept and parse configuration by means of (from highest to lowest
priority) command-line arguments and environment variables.

For more detailed information about the arguments API, please visit the
[Args package GoDoc page](https://godoc.org/github.com/newrelic/infra-integrations-sdk/args). 

## Naming format convention

Referring an argument by its name requires using a different naming format depending on where are you referring the
argument to:

* From **Go code**, an argument is named in `CamelCase` format. Example: `HostName`.
* From the **command-line interface**, an argument is named in lowercase `snake_case` format. Example: `host_name`.
    - This case is also used when the arguments are passed to the integration
      [from a YAML configuration file](../tutorial.md#configuration-of-the-integration-(for-events)).
* From the **environment variables**, an argument is named in `MACRO_CASE` format. Example: `HOST_NAME`.

The `GoSDK v3` arguments API automatically converts between the above kinds of cases.

## Bundled arguments

The arguments API provides a `DefaultArgumentList` struct that contains all the arguments any integration should
consider. All of them are `bool` and default to `false`. They are described following, named by its
[Go code convention](#naming-format-convention):

* `Events`: whether the integration should publish events' data.
* `Inventory`: whether the integration should publish inventory data.
* `Metrics`: whether the integration should publish metrics' data.
* `Pretty`: whether the output JSON must be pretty-formatted.
* `Verbose`: whether the integration should log extra, detailed information.
* `Metadata`: whether the integration should be decorated with `NRI_`prefixed environment provided key-value attributes (ie: `NRI_FOO=BAR`).
* `NriAddHostname`: if true, agent will decorate all the metrics with the `hostname`.
* `NriCluster`: if any value is provided, all the metrics will be decorated with `clusterName: value`. 
* `NriService`: if any value is provided, all the metrics will be decorated with `serviceName: value`. 

An example of

## Adding your own arguments

Your arguments must be defined a struct fields. Such struct should embed the `DefaultArgumentsList` struct, so you
provide a common set of configuration options for all the integrations.

Each struct field should be succeeded by a raw string describing the metadata of the argument, in the form
`` `default:"value" help:"human-readable help description"` ``.

### Example

The next code defines and sets up a set of arguments list that includes the default `GoSDK v3` arguments list
(package info and imports have been omitted for the sake of brevity):

```go
type myArguments struct {
	args.DefaultArgumentList
	SomeInt int `default:"1000" help:"Some integer value"`
	SomeString string `default:"hello" help:"Some string value"`
}

func main() {
	var arguments myArguments

	args.SetupArgs(&arguments)

	fmt.Printf("%+v\n", arguments)
}
```

Assuming this code has been compiled into an executable named `argsTest`, invoking `argsTest -h` should show help
for the users:

```
Usage of argsTest:
  -events
        Publish events data.
  -inventory
        Publish inventory data.
  -metrics
        Publish metrics data.
  -pretty
        Print pretty formatted JSON.
  -some_int int
        Some integer value (default 1000)
  -some_string string
        Some string value (default "hello")
  -verbose
        Print more information to logs.
```

Invoking `argsTest` without arguments should show the default values:

```
{DefaultArgumentList:{Verbose:false Pretty:false Metrics:false Inventory:false
Events:false} SomeInt:1000 SomeString:hello}
```

You can combine command-line with environment arguments, giving priority to command-line:

```
$ export METRICS=true
$ export SOME_INT=123456
$ export SOME_STRING=goodbye
$ go run args.go -some_string "OHAI rules"

{DefaultArgumentList:{Verbose:false Pretty:false Metrics:true Inventory:false
Events:false} SomeInt:123456 SomeString:OHAI rules}
```
