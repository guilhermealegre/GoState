package message

import (
	"fmt"
	"strings"
)

const (
	startedText = "Started"
	errorText   = "ERROR"
)

var (
	stoppedText = fmt.Sprintf("%sStopped%s", ColorRed, ColorReset)
)

// StartMessage writes a start message
func StartMessage(message string) {
	message = strings.ReplaceAll(message, "\n", fmt.Sprintf("%s :: %s\n:: %s", ColorReset, startedText, ColorGreen))
	text := fmt.Sprintf(":: %s%s%s :: %s", ColorGreen, message, ColorReset, startedText)
	fmt.Println(text)
}

// StopMessage writes a stop message
func StopMessage(message string) {
	message = strings.ReplaceAll(message, "\n", fmt.Sprintf("%s :: %s\n:: %s", ColorReset, stoppedText, ColorGreen))
	text := fmt.Sprintf(":: %s%s%s :: %s", ColorGreen, message, ColorReset, stoppedText)
	fmt.Println(text)
}

// Message writes a message
func Message(service string, message string) {
	text := fmt.Sprintf(":: %s%s%s :: %s", ColorGreen, service, ColorReset, message)
	fmt.Println(text)
}

// ErrorMessage writes a error message
func ErrorMessage(message string, err error) {
	text := fmt.Sprintf(":: %s%s%s :: %s", ColorRed, message, ColorReset, errorText)
	if err != nil {
		text = fmt.Sprintf("%s: %s", text, err.Error())
	}
	fmt.Println(text)
}
