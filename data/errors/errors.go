package errors

import (
	"fmt"
)

// ParameterCannotBeEmpty creates an "param cannot be empty" error
func ParameterCannotBeEmpty(param string) error {
	return fmt.Errorf("%s cannot be empty", param)
}
