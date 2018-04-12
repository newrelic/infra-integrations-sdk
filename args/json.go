package args

import (
	"encoding/json"
	"flag"
	"fmt"
)

// JSON type, to be used from the arguments structs.
// This argument type will parse a serialized JSON string into a map which
type JSON struct {
	value interface{}
}

// NewJSON returns a new JSON holder containing the given value
func NewJSON(value interface{}) *JSON {
	return &JSON{value}
}

// Set unmarshals the given string into the JSON holder
func (i *JSON) Set(s string) error {
	if err := json.Unmarshal([]byte(s), &(i.value)); err != nil {
		return fmt.Errorf("Bad JSON, %v", err)
	}
	return nil
}

// Get returns the hold value
func (i *JSON) Get() interface{} { return i.value }

func (i *JSON) String() string {
	s, _ := json.Marshal(&(i.value))
	return string(s)
}

func jsonVar(p *JSON, name string, value string, usage string) {
	flag.CommandLine.Var(p, name, usage)
}
