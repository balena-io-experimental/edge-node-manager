package device

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/radio"
)

type SupportedDevice int

const (
	NRF51822 SupportedDevice = iota
	ESP8266
	MICROBIT
)

func (d SupportedDevice) String() string {
	switch d {
	case NRF51822:
		return "NRF51822"
	case ESP8266:
		return "ESP8266"
	case MICROBIT:
		return "MicroBit"
	}
	return "Not supported"
}

func Create(d SupportedDevice) DeviceInterface {
	switch d {
	case NRF51822:
		return &Nrf51822{
			Device:         &Device{},
			packetSize:     20,
			nameHandle:     0x03,
			identifyHandle: 0x16,
			restartHandle:  0x18,
		}
	case ESP8266:
		return &Esp8266{Device: &Device{}}
	case MICROBIT:
		return &Microbit{Device: &Device{}}
	}

	return nil
}

type DeviceInterface interface {
	String() string
	Serialise() ([]byte, error)
	Deserialise(b []byte) error

	GetDevice() *Device
	Update(application, commit string) error
	Scan() ([]string, error)
	Online() (bool, error)
	Identify() error
	Restart() error
}

type State int

const (
	UPDATING State = iota
	ONLINE
	OFFLINE
	DOWNLOADING
)

type Device struct {
	LocalUUID       string               `json:"localUUID"`
	ApplicationUUID string               `json:"applicationUUID"`
	ResinUUID       string               `json:"resinUUID"`
	Commit          string               `json:"commit"`
	LastSeen        time.Time            `json:"lastSeen, string"`
	State           State                `json:"state, int"`
	Progress        float32              `json:"progress, float32"`
	Radio           radio.RadioInterface `json:"-"`
}

func (d Device) String() string {
	return fmt.Sprintf("Device type: %T"+
		"LocalUUID: %s, "+
		"ApplicationUUID: %s, "+
		"ResinUUID: %s, "+
		"Commit: %s, "+
		"LastSeen: %s, "+
		"State: %s, "+
		"Progress: %f"+
		"Radio type: %T",
		d, //How to make this print actual device type rather than just device.Device
		d.LocalUUID,
		d.ApplicationUUID,
		d.ResinUUID,
		d.Commit,
		d.LastSeen.Format(time.RFC3339),
		d.State,
		d.Progress,
		d.Radio.GetRadio()) //How to make this print actual radio type rather than just radio.Device
}

func (d Device) Serialise() ([]byte, error) {
	b, err := json.Marshal(d)
	return b, err
}

func (d *Device) Deserialise(b []byte) error {
	return json.Unmarshal(b, d)
}
