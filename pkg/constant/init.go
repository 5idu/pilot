package constant

import (
	"os"
	"path/filepath"
)

var (
	appName string
)

const (
	EnvAppName = "APP_NAME"
)

func init() {
	if appName == "" {
		appName = os.Getenv(EnvAppName)
		if appName == "" {
			appName = filepath.Base(os.Args[0])
		}
	}
}

func AppName() string {
	return appName
}
