package radio

import (
	"log"
	"time"
)

type Zigbee struct {
	*Radio
}

func (r Zigbee) GetRadio() *Radio {
	return r.Radio
}

func (r Zigbee) Scan(name string, timeout time.Duration) ([]string, error) {
	log.Println("Scanning ZigBee")
	return make([]string, 1), nil
}

func (r Zigbee) Online(name string, timeout time.Duration) (bool, error) {
	return true, nil
}

func (r Zigbee) SomethingElse() error {
	return nil
}
