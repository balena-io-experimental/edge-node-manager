package radio

import (
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/radio/bluetooth"
	"github.com/josephroberts/edge-node-manager/radio/wifi"
	"github.com/josephroberts/edge-node-manager/radio/zigbee"
)

// Type defines the supported radio types
type Type string

const (
	BLUETOOTH Type = "Bluetooth"
	WIFI           = "WiFi"
	ZIGBEE         = "ZigBee"
)

// Scan calls Scan on the underlying radio type
func (r Type) Scan(name string, timeout time.Duration) (map[string]bool, error) {
	switch r {
	case BLUETOOTH:
		return bluetooth.Scan(name, timeout)
	case WIFI:
		return wifi.Scan(name, timeout)
	case ZIGBEE:
		return zigbee.Scan(name, timeout)
	default:
		return nil, fmt.Errorf("Radio does not exist")
	}
}
