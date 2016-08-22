package radio

import (
	"log"
	"time"

	"github.com/paypal/gatt"
)

type Bluetooth struct {
	*Radio
	gatt gatt.Device
}

var bluetoothChannel = make(chan gatt.Peripheral)

func (r Bluetooth) GetRadio() *Radio {
	return r.Radio
}

func (r Bluetooth) Scan(name string, timeout time.Duration) ([]string, error) {
	log.Println("Scanning Bluetooth")
	r.gatt.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	r.gatt.Init(onStateChanged)

	devices := make([]string, 0, 10)
	for {
		select {
		case <-time.After(timeout * time.Second):
			r.gatt.StopScanning()
			return devices, nil
		case onlineDevice := <-bluetoothChannel:
			if onlineDevice.Name() == name {
				devices = append(devices, onlineDevice.ID())
			}

		}
	}
}

func (r Bluetooth) Online(id string, timeout time.Duration) (bool, error) {
	r.gatt.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	r.gatt.Init(onStateChanged)

	for {
		select {
		case <-time.After(timeout * time.Second):
			r.gatt.StopScanning()
			return false, nil
		case onlineDevice := <-bluetoothChannel:
			if onlineDevice.ID() == id {
				return true, nil
			}

		}
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

func (r Bluetooth) WriteCharacteristic(handle, val string, request bool) error {
	return nil
}

func (r Bluetooth) ReadCharacteristic(handle string) ([]byte, error) {
	return nil, nil
}
