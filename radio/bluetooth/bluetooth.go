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
	Radio         gatt.Device
	periphChannel = make(chan gatt.Peripheral)
)

func Scan(id string, timeout time.Duration) (map[string]bool, error) {
	Radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	if err := Radio.Init(onStateChanged); err != nil {
		return nil, err
	}

	devices := make(map[string]bool)

	for {
		select {
		case <-time.After(timeout * time.Second):
			Radio.StopScanning()
			return devices, nil
		case onlineDevice := <-periphChannel:
			if onlineDevice.Name() == id {
				devices[onlineDevice.ID()] = true
			}
		}
	}
}

func Online(id string, timeout time.Duration) (bool, error) {
	Radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	if err := Radio.Init(onStateChanged); err != nil {
		return false, err
	}

	for {
		select {
		case <-time.After(timeout * time.Second):
			Radio.StopScanning()
			return false, nil
		case onlineDevice := <-periphChannel:
			if onlineDevice.ID() == id {
				Radio.StopScanning()
				return true, nil
			}
		}
	}
}

func init() {
	var err error
	if Radio, err = gatt.NewDevice(option.DefaultClientOptions...); err != nil {
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

func onPeriphDiscovered(periph gatt.Peripheral, adv *gatt.Advertisement, rssi int) {
	periphChannel <- periph
}
