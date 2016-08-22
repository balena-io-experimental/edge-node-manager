package radio

import (
	"time"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

type SupportedRadio string

const (
	BLUETOOTH SupportedRadio = "Bluetooth"
	WIFI      SupportedRadio = "WiFi"
	ZIGBEE    SupportedRadio = "ZigBee"
)

func Create(d SupportedRadio) RadioInterface {
	switch d {
	case BLUETOOTH:
		d, _ := gatt.NewDevice(option.DefaultClientOptions...)
		return &Bluetooth{
			Radio: &Radio{},
			gatt:  d,
		}
	case WIFI:
		//return &Esp8266{Device: &Device{}}
	case ZIGBEE:
		//return &Microbit{Device: &Device{}}
	}

	return nil
}

type RadioInterface interface {
	GetRadio() *Radio
	Scan(name string, timeout time.Duration) ([]string, error)
	Online(id string, timeout time.Duration) (bool, error)
}

type Radio struct {
}
