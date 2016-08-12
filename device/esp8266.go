package device

import "log"

type Esp8266 struct {
	*Device
}

func (d Esp8266) GetDevice() *Device {
	return d.Device
}

func (d Esp8266) Update(application, commit string) error {
	log.Println("Updating ESP8266")
	return nil
}

func (d Esp8266) Scan() ([]string, error) {
	return make([]string, 1), nil
}

func (d Esp8266) Online() (bool, error) {
	return true, nil
}

func (d Esp8266) Identify() error {
	return nil
}

func (d Esp8266) Restart() error {
	return nil
}
