package config

import (
	"os"
	"path"
	"strconv"
	"time"
)

// GetLoopDelay returns the time delay in seconds between each application process loop
func GetLoopDelay() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_CONFIG_LOOP_DELAY", "5"))
	return time.Duration(value), err
}

// GetPersistantDirectory returns the root directory used to store the database and application commits
func GetPersistantDirectory() string {
	return getEnv("ENM_CONFIG_PERSISTENT_DIR", "/data")
}

// GetDbDirectory returns the directory used to store the database
func GetDbDirectory() string {
	return path.Join(GetPersistantDirectory(), getEnv("ENM_CONFIG_DB_DIR", "/database"))
}

// GetProxyPort returns the port used to communicate with the proxy visor
func GetProxyPort() string {
	return getEnv("ENM_CONFIG_PROXY_PORT", "3000")
}

func getEnv(key, fallback string) string {
	result := os.Getenv(key)
	if result == "" {
		result = fallback
	}
	return result
}
