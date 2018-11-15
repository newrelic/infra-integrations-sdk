package args

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DefaultArgumentList includes the minimal set of necessary arguments for an integration.
// If all data flags (Inventory, Metrics and Events) are false, all of them are published.
type DefaultArgumentList struct {
	Verbose        bool   `default:"false" help:"Print more information to logs."`
	Pretty         bool   `default:"false" help:"Print pretty formatted JSON."`
	Metrics        bool   `default:"false" help:"Publish metrics data."`
	Inventory      bool   `default:"false" help:"Publish inventory data."`
	Events         bool   `default:"false" help:"Publish events data."`
	Metadata       bool   `default:"false" help:"Add customer defined key-value attributes to the samples."`
	NriAddHostname bool   `default:"false" help:"Add hostname attribute to the samples."`
	NriCluster     string `default:"" help:"Optional. Cluster name"`
	NriService     string `default:"" help:"Optional. Service name"`
}

// All returns if all data should be published
func (d *DefaultArgumentList) All() bool {
	return !d.Inventory && !d.Metrics && !d.Events
}

// HasMetrics returns if metrics should be published
func (d *DefaultArgumentList) HasMetrics() bool {
	return d.Metrics || d.All()
}

// HasEvents returns if events should be published
func (d *DefaultArgumentList) HasEvents() bool {
	return d.Events || d.All()
}

// HasInventory returns if inventory should be published
func (d *DefaultArgumentList) HasInventory() bool {
	return d.Inventory || d.All()
}

// HTTPClientArgumentList are meant to be used as flags from a custom integrations. With this you could
// send this arguments from the command line.
type HTTPClientArgumentList struct {
	HTTPCaBundleFile string        `default: "" help: "Name of the certificate file"`
	HTTPCaBundleDir  string        `default: "" help: "Path where the certificate exists"`
	HTTPTimeout      time.Duration `default:30 help: "Client http timeout in seconds"`
}

func getArgsFromEnv() func(f *flag.Flag) {
	return func(f *flag.Flag) {
		envName := strings.ToUpper(f.Name)
		if os.Getenv(envName) != "" {
			f.Value.Set(os.Getenv(envName)) // nolint: errcheck
		}
	}
}

var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func underscore(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToLower(strings.Join(a, "_"))
}

// SetupArgs parses a struct's definition and populates the arguments from the
// fields it defines. Each of the fields in the struct can define their defaults
// and help string by using tags:
//
//  type Arguments struct {
//     	DefaultArgumentList
//     	Argument1 bool   `default:"false" help:"This is the help we will print"`
//     	Argument2 int    `default:"1" help:"This is the help we will print"`
//     	Argument3 string `default:"value" help:"This is the help we will print"`
//  }
//
// The fields in the struct will be populated with the values set either from
// the command line or from environment variables.
func SetupArgs(args interface{}) error {
	err := defineFlags(args)
	if err != nil {
		return err
	}

	// Override flags from environment variables with the same name
	flag.VisitAll(getArgsFromEnv())

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return err
	}

	return nil
}

// GetDefaultArgs checks if the arguments interface contains a
// 'DefaultArgumentList' field. If there is no 'DefaultArgumentList' field, it
// returns a DefaultArgumentList with default values.
func GetDefaultArgs(arguments interface{}) *DefaultArgumentList {
	if args, ok := arguments.(*DefaultArgumentList); ok {
		return args
	}

	defaultArgsI := reflect.ValueOf(arguments).Elem().FieldByName("DefaultArgumentList")

	if defaultArgsI.IsValid() {
		defaultArgs := defaultArgsI.Addr().Interface().(*DefaultArgumentList)
		return defaultArgs
	}
	return &DefaultArgumentList{}
}

func defineFlags(args interface{}) error {
	val := reflect.ValueOf(args).Elem()

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		// The argument will take the field's name in underscore
		argName := underscore(typeField.Name)
		// We get a generic pointer to the field
		argDefault := valueField.Addr().Interface()
		// Get the default and help tags fom the struct field
		defaultValue := tag.Get("default")
		helpValue := tag.Get("help")

		switch argDefault := argDefault.(type) {
		case *int:
			intVal, err := strconv.Atoi(defaultValue)
			if err != nil {
				return fmt.Errorf("can't parse %s: not an integer", argName)
			}
			flag.IntVar(argDefault, argName, intVal, helpValue)
		case *bool:
			boolVal, err := strconv.ParseBool(defaultValue)
			if err != nil {
				return fmt.Errorf("can't parse %s: not a boolean", argName)
			}
			flag.BoolVar(argDefault, argName, boolVal, helpValue)
		case *string:
			flag.StringVar(argDefault, argName, defaultValue, helpValue)
		case *JSON:
			jsonVar(argDefault, argName, helpValue)
		case *DefaultArgumentList:
			err := defineFlags(argDefault)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("can't parse %s: unsupported type", argName)
		}
	}
	return nil
}
