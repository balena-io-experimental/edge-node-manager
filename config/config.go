package config

import (
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
)

// GetLogLevel returns the log level
func GetLogLevel() log.Level {
	level := getEnv("LOG_LEVEL", "")

	switch level {
	case "DEBUG":
		return log.DebugLevel
	case "INFO":
		return log.InfoLevel
	case "WARN":
		return log.WarnLevel
	case "ERROR":
		return log.ErrorLevel
	}

	return log.DebugLevel
}

// GetLoopDelay returns the time delay in seconds between each application process loop
func GetLoopDelay() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_CONFIG_LOOP_DELAY", "5"))
	return time.Duration(value), err
}

// GetAssetsDir returns the root directory used to store the database and application commits
func GetAssetsDir() string {
	return getEnv("ENM_ASSETS_DIRECTORY", "/data/assets")
}

// GetDbDir returns the directory used to store the database
func GetDbDir() string {
	return getEnv("ENM_DB_DIRECTORY", "/data/database")
}

// GetDbName returns the directory used to store the database
func GetDbName() string {
	return getEnv("ENM_DB_NAME", "my.db")
}

// GetENMAddr returns the address used to serve the API to the supervisor
func GetENMAddr() string {
	return getEnv("RESIN_DEPENDENT_DEVICES_HOOK_ADDRESS", "http://127.0.0.1:3000/v1/devices/")
}

// GetSuperAddr returns the address used to communicate with the supervisor
func GetSuperAddr() string {
	return getEnv("RESIN_SUPERVISOR_ADDRESS", "http://127.0.0.1:4000")
}

// GetSuperAPIKey returns the API key used to communicate with the supervisor
func GetSuperAPIKey() string {
	return getEnv("RESIN_SUPERVISOR_API_KEY", "")
}

// GetSuperAPIVer returns the API key used to communicate with the supervisor
func GetSuperAPIVer() string {
	return getEnv("RESIN_SUPERVISOR_API_VERSION", "v1")
}

func getEnv(key, fallback string) string {
	result := os.Getenv(key)
	if result == "" {
		result = fallback
	}
	return result
}
