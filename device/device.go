package device

import (
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/firmware"
	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/radio"
)

// Type contains the micro and radio that make up a device type
type Type struct {
	Micro micro.Type `json:"Micro"`
	Radio radio.Type `json:"Radio"`
}

// State defines the device states
type State string

const (
	UPDATING    State = "Updating"
	ONLINE            = "Online"
	OFFLINE           = "Offline"
	DOWNLOADING       = "Downloading"
)

// Device contains all the variables needed to define a device
type Device struct {
	Type            `json:"Type"`
	LocalUUID       string    `json:"LocalUUID"`
	ApplicationUUID string    `json:"ApplicationUUID"`
	ResinUUID       string    `json:"ResinUUID"`
	DatabaseUUID    int       `json:"DatabaseUUID"`
	Commit          string    `json:"Commit"`
	LastSeen        time.Time `json:"LastSeen"`
	State           State     `json:"State"`
	Progress        float32   `json:"Progress"`
}

// Interface defines the common functions a device must implement
type Interface interface {
	String() string
	Update(firmware firmware.Firmware) error
	Online() (bool, error)
	Identify() error
	Restart() error
}

func (d Device) String() string {
	return fmt.Sprintf(
		"Micro type: %s, "+
			"Radio type: %s, "+
			"LocalUUID: %s, "+
			"ApplicationUUID: %s, "+
			"ResinUUID: %s, "+
			"DatabaseUUID: %d, "+
			"Commit: %s, "+
			"LastSeen: %s, "+
			"State: %s, "+
			"Progress: %2.2f",
		d.Type.Micro,
		d.Type.Radio,
		d.LocalUUID,
		d.ApplicationUUID,
		d.ResinUUID,
		d.DatabaseUUID,
		d.Commit,
		d.LastSeen,
		d.State,
		d.Progress)
}

// Cast converts the base device type to its actual type
func (d Device) Cast() Interface {
	switch d.Type.Micro {
	case micro.NRF51822:
		return (Nrf51822)(d)
	case micro.ESP8266:
		return (Esp8266)(d)
	}

	return nil

}
