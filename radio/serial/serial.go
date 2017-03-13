package serial

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/config"
)

func Scan(id string, timeout time.Duration) (map[string]bool, error) {
	devices := make(map[string]bool)

	return devices, nil
}

func Online(id string, timeout time.Duration) (bool, error) {
	online := false

	return online, nil
}

func init() {
	log.SetLevel(config.GetLogLevel())

	log.Debug("Created a new serial device")
}
