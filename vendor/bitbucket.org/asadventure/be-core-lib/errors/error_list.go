package errors

import (
	"reflect"
	"strings"
)

// IsEmpty verify if has any error
func (e ErrorDetailsList) IsEmpty() bool {
	return reflect.DeepEqual(e, ErrorDetailsList{})
}

// Error error method
func (e ErrorDetailsList) Error() string {
	var errs string

	for _, err := range e {
		errs = strings.Join([]string{errs, err.Error()}, "; ")
	}

	return errs
}

// NewErrorDetailsList creates a new error details
func NewErrorDetailsList() ErrorDetailsList {
	return ErrorDetailsList{}
}

// Add add error
func (e ErrorDetailsList) Add(err ErrorDetails) ErrorDetailsList {
	e = append(e, err)
	return e
}
