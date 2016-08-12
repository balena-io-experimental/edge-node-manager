package radio

import "time"

type SupportedRadio int

const (
	BLUETOOTH SupportedRadio = iota
	WIFI
	ZIGBEE
)

type RadioInterface interface {
	GetRadio() *Radio
	Scan(name string, timeout time.Duration) ([]string, error)
	Online(id string, timeout time.Duration) (bool, error)
}

type Radio struct {
}
