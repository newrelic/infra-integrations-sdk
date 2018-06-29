# Internal logging

To avoid depending on third-party logging solutions, the SDK provides a very simple `Logger` interface allows injecting
any logger implementation that fulfills the interface:

```go
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}
```

, where the `-f` suffix means that the string and arguments are passed in C-like printf format
(as in [fmt.Printf](https://godoc.org/fmt#Printf)).

Some popular logging solutions (e.g. [Logrus](https://github.com/sirupsen/logrus)) already implement the above interface,
so their implemented loggers can be used out of the box.

If you don't want to add more dependencies, the SDK provides two functions to instantiate bundled, simple loggers:
[NewStdErr](https://godoc.org/github.com/newrelic/infra-integrations-sdk/log#New), which creates a log whose output
is sent to standard error; and [New](https://godoc.org/github.com/newrelic/infra-integrations-sdk/log#New), which create
a log whose output is written in the `io.Writer` passed as argument.

Thanks to the simplicity of the `Logger` interface, it's easy to implement a custom logging solution. For example:

```go
type verySimpleLogger struct{}

func (s verySimpleLogger) Debugf(format string, args ...interface{}) {
	fmt.Println("DEBUG:", fmt.Sprintf(format, args...))
}

func (s verySimpleLogger) Infof(format string, args ...interface{}) {
	fmt.Println("INFO:", fmt.Sprintf(format, args...))
}

func (s verySimpleLogger) Warnf(format string, args ...interface{}) {
	fmt.Println("WARN:", fmt.Sprintf(format, args...))
}

func (s verySimpleLogger) Errorf(format string, args ...interface{}) {
	fmt.Println("ERROR:", fmt.Sprintf(format, args...))
}
```

Or just a Null logger that would hide all the messages:

```go
type nullLogger struct{}

func (nullLogger) Debugf(format string, args ...interface{}) { }
func (nullLogger) Infof(format string, args ...interface{}) { }
func (nullLogger) Warnf(format string, args ...interface{}) { }
func (nullLogger) Errorf(format string, args ...interface{}) { }
```
