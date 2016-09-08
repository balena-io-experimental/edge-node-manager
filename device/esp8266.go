package device

import (
	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/radio/wifi"
)

type Esp8266 Device

func (d Esp8266) String() string {
	return (Device)(d).String()
}

func (d Esp8266) Update(application, commit string) error {
	log.WithFields(log.Fields{
		"Device":           d,
		"Application UUID": application,
		"Commit":           commit,
	}).Debug("Update")
	return nil
}

func (d Esp8266) Online() (bool, error) {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Online")
	return wifi.Online(d.LocalUUID, 10)
}

func (d Esp8266) Identify() error {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Identify")
	return wifi.Post(d.LocalUUID, "hey")
}

func (d Esp8266) Restart() error {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Restart")
	return wifi.Post(d.LocalUUID, "hey again")
}
