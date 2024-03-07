package config

import (
	"fmt"

	"bitbucket.org/asadventure/be-core-lib/errors"
	"bitbucket.org/asadventure/be-infrastructure-lib/config"
	"bitbucket.org/asadventure/be-infrastructure-lib/domain/message"
)

// ConfigFile the configuration file
const configFile = "error.yaml"

// cacheError
var cacheError *SError

// SError
type SError struct {
	Code string `json:"code"`
}

// initConfigs initialize the configurations
func init() {
	if err := config.Load(cacheError.ConfigFile(), &cacheError); err != nil {
		err = fmt.Errorf("Error loading config file %s: %s", cacheError.ConfigFile(), err)
		message.ErrorMessage("errors", err)
		fmt.Println("Error while loading error configuration file: ", err)
	}
}

// ConfigFile gets the configuration file
func (e *SError) ConfigFile() string {
	return configFile
}

// GetError creates an error details
func GetError(code string, msg string, level errors.Level, opts ...errors.Opt) func() errors.ErrorDetails {
	return func() errors.ErrorDetails {
		if cacheError == nil {
			cacheError = &SError{
				Code: "",
			}
		}
		return errors.NewErrorDetails(cacheError.Code+"-"+code, msg, level, opts...)
	}
}
