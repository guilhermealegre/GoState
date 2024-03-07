package errors

import (
	"fmt"
)

// IsEmpty verify if has any error
func (e ErrorDetails) IsEmpty() bool {
	return e == ErrorDetails{}
}

// Formats formats an error message
func (e ErrorDetails) Formats(values ...interface{}) ErrorDetails {
	return ErrorDetails{
		Level:      e.Level,
		Code:       e.Code,
		Field:      e.Field,
		StatusCode: e.StatusCode,
		ErrorMsg:   fmt.Sprintf(e.ErrorMsg, values...),
	}
}

// Error error method
func (e ErrorDetails) Error() string {
	return e.ErrorMsg
}

// GetLevel get level
func (e ErrorDetails) GetLevel() Level {
	return e.Level
}

// GetField get field
func (e ErrorDetails) GetField() string {
	return e.Field
}

// GetStatusCode Get status code
func (e ErrorDetails) GetStatusCode() int {
	return e.StatusCode
}

func (e ErrorDetails) SetField(field string) ErrorDetails {
	e.Field = field
	return e
}

func (e ErrorDetails) SetStatusCode(statusCode int) ErrorDetails {
	e.StatusCode = statusCode
	return e
}

// NewErrorDetails creates a new error details
func NewErrorDetails(code string, msg string, level Level, opts ...Opt) ErrorDetails {
	errorDetails := ErrorDetails{
		Level:    level,
		Code:     code,
		ErrorMsg: msg,
	}

	if len(opts) > 0 {
		for _, opt := range opts {
			switch opt.GetKey() {
			case OptStatusCode:
				errorDetails.StatusCode = opt.Value.(int)
			}
		}
	}

	return errorDetails
}
