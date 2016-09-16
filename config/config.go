package config

import (
	"os"
	"strconv"
	"time"
)

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

// GetSuperAddr returns the address used to communicate with the supervisor
func GetSuperAddr() string {
	return getEnv("RESIN_SUPERVISOR_ADDRESS", "http://localhost:3000")
}

// GetSuperAPIKey returns the API key used to communicate with the supervisor
func GetSuperAPIKey() string {
	return getEnv("RESIN_SUPERVISOR_API_KEY", "test")
}

// GetSuperAPIVer returns the API key used to communicate with the supervisor
func GetSuperAPIVer() string {
	return getEnv("RESIN_SUPERVISOR_API_KEY", "v1")
}

func getEnv(key, fallback string) string {
	result := os.Getenv(key)
	if result == "" {
		result = fallback
	}
	return result
}
