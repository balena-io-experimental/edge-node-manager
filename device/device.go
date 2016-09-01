package device

import (
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/firmware"
	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/radio"
)

type DeviceType struct {
	Micro micro.MicroType `json:"Micro"`
	Radio radio.RadioType `json:"Radio"`
}

type State string

const (
	UPDATING    State = "Updating"
	ONLINE            = "Online"
	OFFLINE           = "Offline"
	DOWNLOADING       = "Downloading"
)

type Device struct {
	DeviceType      `json:"DeviceType"`
	LocalUUID       string    `json:"LocalUUID"`
	ApplicationUUID string    `json:"ApplicationUUID"`
	ResinUUID       string    `json:"ResinUUID"`
	DatabaseUUID    int       `json:"DatabaseUUID"`
	Commit          string    `json:"Commit"`
	LastSeen        time.Time `json:"LastSeen"`
	State           State     `json:"State"`
	Progress        float32   `json:"Progress"`
}

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
		d.DeviceType.Micro,
		d.DeviceType.Radio,
		d.LocalUUID,
		d.ApplicationUUID,
		d.ResinUUID,
		d.DatabaseUUID,
		d.Commit,
		d.LastSeen,
		d.State,
		d.Progress)
}

func (d Device) ChooseType() Interface {
	switch d.DeviceType.Micro {
	case micro.NRF51822:
		return (Nrf51822)(d)
	case micro.ESP8266:
		return (Esp8266)(d)
	}

	return nil

}
