package radio

import (
	"log"
	"time"
)

type Wifi struct {
	*Radio
}

func (r Wifi) GetRadio() *Radio {
	return r.Radio
}

func (r Wifi) Scan(name string, timeout time.Duration) ([]string, error) {
	log.Println("Scanning Wifi")
	return make([]string, 1), nil
}

func (r Wifi) Online(name string, timeout time.Duration) (bool, error) {
	return true, nil
}

func (r Wifi) Post(url, data string) error {
	return nil
}
