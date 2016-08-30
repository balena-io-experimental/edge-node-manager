package bluetooth

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

/*
 * Uses the gatt package
 * https://github.com/paypal/gatt
 */
var (
	radio            gatt.Device
	bluetoothChannel = make(chan gatt.Peripheral)
)

func Scan(name string, timeout time.Duration) (map[string]bool, error) {
	radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	if err := radio.Init(onStateChanged); err != nil {
		return nil, err
	}

	devices := make(map[string]bool)

	for {
		select {
		case <-time.After(timeout * time.Second):
			radio.StopScanning()
			return devices, nil
		case onlineDevice := <-bluetoothChannel:
			if onlineDevice.Name() == name {
				devices[onlineDevice.ID()] = true
			}
		}
	}
}

func Online(id string, timeout time.Duration) (bool, error) {
	radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	if err := radio.Init(onStateChanged); err != nil {
		return false, err
	}

	for {
		select {
		case <-time.After(timeout * time.Second):
			radio.StopScanning()
			return false, nil
		case onlineDevice := <-bluetoothChannel:
			if onlineDevice.ID() == id {
				return true, nil
			}
		}
	}
}

func WriteCharacteristic(handle, val byte, request bool) error {
	return nil
}

func ReadCharacteristic(handle byte) ([]byte, error) {
	return nil, nil
}

func init() {
	var err error
	if radio, err = gatt.NewDevice(option.DefaultClientOptions...); err != nil {
		log.WithFields(log.Fields{
			"Options": option.DefaultClientOptions,
			"Error":   err,
		}).Fatal("Unable to create a new gatt device")
	}

	log.WithFields(log.Fields{
		"Options": option.DefaultClientOptions,
	}).Debug("Created new gatt device")
}

func onStateChanged(device gatt.Device, state gatt.State) {
	switch state {
	case gatt.StatePoweredOn:
		device.Scan([]gatt.UUID{}, false)
	default:
		device.StopScanning()
	}
}

func onPeriphDiscovered(peripheral gatt.Peripheral, advertisment *gatt.Advertisement, rssi int) {
	bluetoothChannel <- peripheral
}
