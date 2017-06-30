package args

import (
	"encoding/json"
	"flag"
	"fmt"
)

// JSON type, to be used from the arguments structs
// This argument type will parse a serialized JSON string into a map which
type JSON map[string]interface{}

func (i *JSON) Set(s string) error {
	if err := json.Unmarshal([]byte(s), i); err != nil {
		return fmt.Errorf("Bad JSON, %v", err)
	}
	return nil
}

func (i *JSON) Get() interface{} { return *i }

func (i *JSON) String() string {
	s, _ := json.Marshal(i)
	return string(s)
}

func jsonVar(p *JSON, name string, value string, usage string) {
	flag.CommandLine.Var(p, name, usage)
}
