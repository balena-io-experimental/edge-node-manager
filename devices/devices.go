package devices

import (
	"encoding/json"

	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
)

// Put puts all devices for a specific application into the database
func Put(applicationUUID int, devices map[string]*device.Device) error {
	buffer := make(map[string][]byte)
	for _, value := range devices {
		bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		buffer[value.UUID] = bytes
	}

	return database.PutDevices(applicationUUID, buffer)
}

// Get gets all devices for a specific application
func Get(applicationUUID int) (map[string]*device.Device, error) {
	buffer, err := database.GetDevices(applicationUUID)
	if err != nil {
		return nil, err
	}

	devices := make(map[string]*device.Device)
	for _, value := range buffer {
		var device device.Device
		if err = json.Unmarshal(value, &device); err != nil {
			return nil, err
		}

		devices[device.LocalUUID] = &device
	}

	return devices, nil
}
