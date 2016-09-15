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

// GetResinSupervisorAddress returns the address used to communicate with the supervisor
func GetResinSupervisorAddress() string {
	return getEnv("RESIN_SUPERVISOR_ADDRESS", "http://localhost:3000")
}

// GetResinSupervisorAPIKey returns the API key used to communicate with the supervisor
func GetResinSupervisorAPIKey() string {
	return getEnv("RESIN_SUPERVISOR_API_KEY", "test")
}

func getEnv(key, fallback string) string {
	result := os.Getenv(key)
	if result == "" {
		result = fallback
	}
	return result
}
