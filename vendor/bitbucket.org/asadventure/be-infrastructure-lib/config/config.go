package config

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	// basePath the base path for all the configurations
	basePath = "conf/"
)

// Load loads a configuration file from the path basePath to the obj struct
func Load(file string, obj interface{}) (err error) {
	viper.SetConfigFile(getCwd() + path.Join(basePath, file))

	if err = viper.ReadInConfig(); err != nil {
		return err
	}

	if err = viper.Unmarshal(obj); err != nil {
		return err
	}

	return nil
}

func getCwd() string {
	var cwd string

	if isTestRunning() {
		x, _ := filepath.Abs("./")
		aDir := strings.Split(x, "/")
		cwd = strings.Join(aDir[0:len(aDir)-4], "/") + "/"
	}

	return cwd
}
func isTestRunning() bool {
	return strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/_test/")
}
