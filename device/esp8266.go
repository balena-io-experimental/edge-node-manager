package device

import (
	"log"

	"github.com/josephroberts/edge-node-manager/radio/wifi"
)

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

func (d Esp8266) Online() (bool, error) {
	log.Println("Checking if ESP8266 online")
	return wifi.Online(d.LocalUUID, 10)
}

func (d Esp8266) Identify() error {
	log.Println("Identifying ESP8266")
	return wifi.Post(d.LocalUUID, "hey")
}

func (d Esp8266) Restart() error {
	log.Println("Restarting ESP8266")
	return wifi.Post(d.LocalUUID, "hey again")
}
