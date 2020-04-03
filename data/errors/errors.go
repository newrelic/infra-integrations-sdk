package errors

import (
	"fmt"
)

// ErrParameterCannotBeEmpty creates an "param cannot be empty" error
func ErrParameterCannotBeEmpty(param string) error {
	return fmt.Errorf("%s cannot be empty", param)
}
