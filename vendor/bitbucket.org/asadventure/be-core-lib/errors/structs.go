package errors

// Level error level
type Level int

// Optional
type Optional string

// Opts
type Opt struct {
	Key   Optional `json:"key"`
	Value any      `json:"value"`
}

func (o Opt) GetKey() Optional {
	return o.Key
}

func (o Opt) GetValue() any {
	return o.Value
}

// ErrorDetailsList error details list
type ErrorDetailsList []ErrorDetails

// ErrorDetails error details
type ErrorDetails struct {
	// Level
	Level Level `json:"level"`
	// Code
	Code string `json:"code"`
	// Error Message
	ErrorMsg string `json:"error"`
	// Field
	Field string `json:"field,omitempty"`
	// Status Code HTTP error to be returned
	StatusCode int `json:"statusCode,omitempty"`
}
