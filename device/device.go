package device

import (
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/device/deviceType"
	"github.com/josephroberts/edge-node-manager/radio/bluetooth"
	"github.com/josephroberts/edge-node-manager/radio/radioType"
	"github.com/josephroberts/edge-node-manager/radio/wifi"
	"github.com/josephroberts/edge-node-manager/radio/zigbee"
)

type Type struct {
	Device deviceType.Type
	Radio  radioType.Type
}

func (t Type) Scan(name string, timeout time.Duration) ([]string, error) {
	switch t.Radio {
	case radioType.BLUETOOTH:
		return bluetooth.Scan(name, timeout)
	case radioType.WIFI:
		return wifi.Scan(name, timeout)
	case radioType.ZIGBEE:
		return zigbee.Scan(name, timeout)
	default:
		return nil, fmt.Errorf("Radio %s does not exist", t.Radio)
	}
}

func Create(t *Type) Interface {
	switch t.Device {
	case deviceType.NRF51822:
		return &Nrf51822{
			Device: &Device{
				Type: t,
			},
		}
	case deviceType.ESP8266:
		return &Esp8266{
			Device: &Device{
				Type: t,
			},
		}
	}

	return nil
}

func Provision(t *Type, localUUID, applicationUUID, resinUUID string) Interface {
	device := Create(t)
	device.GetDevice().LocalUUID = localUUID
	device.GetDevice().ApplicationUUID = applicationUUID
	device.GetDevice().LastSeen = time.Now()
	device.GetDevice().State = ONLINE
	return device
}

type Interface interface {
	String() string

	GetDevice() *Device
	Update(application, commit string) error
	Online() (bool, error)
	Identify() error
	Restart() error
}

type State string

const (
	UPDATING    State = "Updating"
	ONLINE            = "Online"
	OFFLINE           = "Offline"
	DOWNLOADING       = "Downloading"
)

type Device struct {
	*Type
	LocalUUID       string
	ApplicationUUID string
	ResinUUID       string
	Commit          string
	LastSeen        time.Time
	State           State
	Progress        float32
}

func (d Device) String() string {
	return fmt.Sprintf(
		"Device type: %s, "+
			"Radio type: %s, "+
			"LocalUUID: %s, "+
			"ApplicationUUID: %s, "+
			"ResinUUID: %s, "+
			"Commit: %s, "+
			"LastSeen: %s, "+
			"State: %s, "+
			"Progress: %.2f",
		d.Type.Device,
		d.Type.Radio,
		d.LocalUUID,
		d.ApplicationUUID,
		d.ResinUUID,
		d.Commit,
		d.LastSeen.Format(time.RFC3339),
		d.State,
		d.Progress)
}
