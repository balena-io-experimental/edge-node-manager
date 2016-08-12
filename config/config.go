package config

import (
	"os"
	"path"
	"strconv"
	"time"
)

func GetLoopDelay() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_CONFIG_LOOP_DELAY", "5"))
	return time.Duration(value), err
}

func GetPersistantDirectory() string {
	return getEnv("ENM_CONFIG_PERSISTANT_DIR", "/data")
}

func GetDbDirectory() string {
	return path.Join(GetPersistantDirectory(), getEnv("ENM_CONFIG_DB_DIR", "/database"))
}
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
