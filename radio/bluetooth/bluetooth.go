package bluetooth

import (
	"log"
	"time"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

/*
Uses the gatt package
https://github.com/paypal/gatt
*/

var radio gatt.Device
var bluetoothChannel = make(chan gatt.Peripheral)

func Scan(name string, timeout time.Duration) ([]string, error) {
	log.Printf("Scanning for bluetooth devices named %s\r\n", name)
	initialise()
	radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	radio.Init(onStateChanged)

	devices := make([]string, 0, 10)

	for {
		select {
		case <-time.After(timeout * time.Second):
			radio.StopScanning()
			return devices, nil
		case onlineDevice := <-bluetoothChannel:
			if onlineDevice.Name() == name {
				devices = append(devices, onlineDevice.ID())
			}
		}
	}
}

func Online(id string, timeout time.Duration) (bool, error) {
	log.Printf("Checking if bluetooth device %s is online\r\n", id)
	initialise()
	radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	radio.Init(onStateChanged)

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

func initialise() {
	if radio == nil {
		var err error
		if radio, err = gatt.NewDevice(option.DefaultClientOptions...); err != nil {
			log.Fatalf("Unable to create a new gatt device: %v", err)
		}
		log.Println("Created a new gatt device")
	}
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
