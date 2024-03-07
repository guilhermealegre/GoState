package errors

// Error Levels
const (
	// Fatal
	Fatal Level = iota
	// Error
	Error
	// Warning
	Warning
	// Info
	Info
	// Debug
	Debug
)

// Error optional parameters
const (
	OptStatusCode Optional = "StatusCode"
)
