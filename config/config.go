package config

import (
	"os"
	"path"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
)

// GetLogLevel returns the log level
func GetLogLevel() log.Level {
	return switchLogLevel(getEnv("ENM_LOG_LEVEL", ""))
}

// GetDependentLogLevel returns the log level for dependent devices
func GetDependentLogLevel() log.Level {
	return switchLogLevel(getEnv("DEPENDENT_LOG_LEVEL", ""))
}

// GetSupervisorCheckDelay returns the time delay in seconds between each supervisor check at startup
func GetSupervisorCheckDelay() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_SUPERVISOR_CHECK_DELAY", "1"))
	return time.Duration(value) * time.Second, err
}

// GetLoopDelay returns the time delay in seconds between each application process loop
func GetLoopDelay() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_CONFIG_LOOP_DELAY", "10"))
	return time.Duration(value) * time.Second, err
}

// GetPauseDelay returns the time delay in seconds between each pause check
func GetPauseDelay() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_CONFIG_PAUSE_DELAY", "10"))
	return time.Duration(value) * time.Second, err
}

// GetHotspotSSID returns the SSID to be used for the hotspot
func GetHotspotSSID() string {
	return getEnv("ENM_HOTSPOT_SSID", "resin-hotspot")
}

// GetHotspotPassword returns the password to be used for the hotspot
func GetHotspotPassword() string {
	return getEnv("ENM_HOTSPOT_PASSWORD", "resin-hotspot")
}

// GetShortBluetoothTimeout returns the timeout for each instantaneous bluetooth operation
func GetShortBluetoothTimeout() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_BLUETOOTH_SHORT_TIMEOUT", "1"))
	return time.Duration(value) * time.Second, err
}

// GetLongBluetoothTimeout returns the timeout for each long running bluetooth operation
func GetLongBluetoothTimeout() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_BLUETOOTH_LONG_TIMEOUT", "10"))
	return time.Duration(value) * time.Second, err
}

// GetAvahiTimeout returns the timeout for each Avahi scan operation
func GetAvahiTimeout() (time.Duration, error) {
	value, err := strconv.Atoi(getEnv("ENM_AVAHI_TIMEOUT", "10"))
	return time.Duration(value) * time.Second, err
}

// GetUpdateRetries returns the number of times the firmware update process should be attempted
func GetUpdateRetries() (int, error) {
	return strconv.Atoi(getEnv("ENM_UPDATE_RETRIES", "1"))
}

// GetAssetsDir returns the root directory used to store the database and application commits
func GetAssetsDir() string {
	return getEnv("ENM_ASSETS_DIRECTORY", "/data/assets")
}

// GetDbDir returns the directory used to store the database
func GetDbDir() string {
	return getEnv("ENM_DB_DIRECTORY", "/data/database")
}

// GetDbPath returns the path used to store the database
func GetDbPath() string {
	directory := GetDbDir()
	file := getEnv("ENM_DB_FILE", "enm.db")

	return path.Join(directory, file)
}

// GetVersion returns the API version used to communicate with the supervisor
func GetVersion() string {
	return getEnv("ENM_API_VERSION", "v1")
}

// GetSuperAddr returns the address used to communicate with the supervisor
func GetSuperAddr() string {
	return getEnv("RESIN_SUPERVISOR_ADDRESS", "http://127.0.0.1:4000")
}

// GetSuperAPIKey returns the API key used to communicate with the supervisor
func GetSuperAPIKey() string {
	return getEnv("RESIN_SUPERVISOR_API_KEY", "")
}

// GetLockFileLocation returns the location of the lock file
func GetLockFileLocation() string {
	return getEnv("ENM_LOCK_FILE_LOCATION", "/tmp/resin/resin-updates.lock")
}

func getEnv(key, fallback string) string {
	result := os.Getenv(key)
	if result == "" {
		result = fallback
	}
	return result
}

func switchLogLevel(level string) log.Level {
	switch level {
	case "Debug":
		return log.DebugLevel
	case "Info":
		return log.InfoLevel
	case "Warn":
		return log.WarnLevel
	case "Error":
		return log.ErrorLevel
	case "Fatal":
		return log.FatalLevel
	case "Panic":
		return log.PanicLevel
	}

	return log.InfoLevel
}
