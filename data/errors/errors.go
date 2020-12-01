package errors

import (
	"fmt"
)

// ParameterCannotBeEmpty creates an "param cannot be empty" error
func ParameterCannotBeEmpty(param string) error {
	return fmt.Errorf("%s cannot be empty", param)
}

// ParameterCannotBeNegative creates an "param cannot be negative" error
func ParameterCannotBeNegative(param string, value interface{}) error {
	return fmt.Errorf("%s (%v) cannot be negative. ", param, value)
}
