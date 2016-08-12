package device

import "log"

type Microbit struct {
	*Device
}

func (d Microbit) GetDevice() *Device {
	return d.Device
}

func (d Microbit) Update(application, commit string) error {
	log.Println("Updating Microbit")
	return nil
}

func (d Microbit) Scan() ([]string, error) {
	return make([]string, 1), nil
}

func (d Microbit) Online() (bool, error) {
	return true, nil
}

func (d Microbit) Identify() error {
	return nil
}

func (d Microbit) Restart() error {
	return nil
}
