package device

import (
	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/radio/wifi"
)

// Esp8266 is an ESP8266 based device
// https://en.wikipedia.org/wiki/ESP8266
type Esp8266 Device

func (d Esp8266) String() string {
	return (Device)(d).String()
}

// Update updates the device following the firmware-over-the-air update process
func (d Esp8266) Update(path string) error {
	log.WithFields(log.Fields{
		"Device": d,
		"Path":   path,
	}).Debug("Update")
	return nil
}

// Online checks whether the device is online
func (d Esp8266) Online() (bool, error) {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Online")
	return wifi.Online(d.LocalUUID, 10)
}

// Identify flashes LEDs' on the device
func (d Esp8266) Identify() error {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Identify")
	return nil
}

// Restart restarts the device
func (d Esp8266) Restart() error {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Restart")
	return nil
}
