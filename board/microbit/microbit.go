package microbit

import (
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/micro/nrf51822"
	"github.com/josephroberts/edge-node-manager/radio/bluetooth"
	"github.com/paypal/gatt"
)

type Microbit struct {
	Micro nrf51822.Nrf51822
}

func (b Microbit) Update(path string) error {
	if err := b.Micro.ExtractFirmware(path, "micro-bit.bin", "micro-bit.dat"); err != nil {
		return err
	}

	bluetooth.Radio.Handle(
		gatt.PeripheralDiscovered(b.Micro.OnPeriphDiscovered),
		gatt.PeripheralConnected(b.bootloadOnPeriphConnected),
		gatt.PeripheralDisconnected(b.Micro.OnPeriphDisconnected),
	)

	if err := bluetooth.Radio.Init(bluetooth.OnStateChanged); err != nil {
		return err
	}

	var savedErr error
	for {
		select {
		case savedErr = <-b.Micro.ErrChannel:
		case state := <-b.Micro.StateChannel:
			if state["connected"] {
				log.Debug("Connected")
			} else {
				log.Debug("Disconnected")

				if !state["restart"] {
					return savedErr
				}

				bluetooth.Radio.Handle(
					gatt.PeripheralDiscovered(b.Micro.OnPeriphDiscovered),
					gatt.PeripheralConnected(b.Micro.UpdateOnPeriphConnected),
					gatt.PeripheralDisconnected(b.Micro.OnPeriphDisconnected),
				)
				if err := bluetooth.Radio.Init(bluetooth.OnStateChanged); err != nil {
					return err
				}
			}
		}
	}
}

func (b Microbit) Scan(applicationUUID int) (map[string]bool, error) {
	id := "BBC micro:bit [" + strconv.Itoa(applicationUUID) + "]"
	return bluetooth.Scan(id, 10)
}

func (b Microbit) Online() (bool, error) {
	return bluetooth.Online(b.Micro.LocalUUID, 10)
}

func (b Microbit) Restart() error {
	return b.Micro.ProcessRequest(b.restartOnPeriphConnected)
}

func (b Microbit) Identify() error {
	return b.Micro.ProcessRequest(b.identifyOnPeriphConnected)
}

func (b Microbit) bootloadOnPeriphConnected(periph gatt.Peripheral, err error) {
	defer periph.Device().CancelConnection(periph)

	b.Micro.StateChannel <- map[string]bool{
		"connected": true,
	}

	if err := periph.SetMTU(500); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	if err := b.startBootloader(periph); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	// Time delay to allow device restart after being placed into bootloader mode
	time.Sleep(1 * time.Second)
}

func (b Microbit) startBootloader(periph gatt.Peripheral) error {
	log.Debug("Starting bootloader mode")

	name, err := bluetooth.GetName(periph)
	if err != nil {
		return err
	}

	// The device name is used to check whether the device is in bootloader mode
	if name == "DfuTarg" {
		log.Debug("In bootloader mode")
		b.Micro.StateChannel <- map[string]bool{
			"restart": false,
		}
		return nil
	}

	b.Micro.StateChannel <- map[string]bool{
		"restart": true,
	}

	characteristic, err := bluetooth.GetChar("e95d93b0251d470aa062fa1922dfa9a8", "e95d93b1251d470aa062fa1922dfa9a8", "2902", gatt.CharRead+gatt.CharWrite, 13, 14)
	if err != nil {
		return err
	}

	if err = periph.WriteCharacteristic(characteristic, []byte{nrf51822.Start}, false); err != nil {
		return err
	}

	log.Debug("Started bootloader mode")

	return nil
}

func (b Microbit) restartOnPeriphConnected(periph gatt.Peripheral, err error) {
	defer periph.Device().CancelConnection(periph)

	b.Micro.StateChannel <- map[string]bool{
		"connected": true,
	}

	characteristic, err := bluetooth.GetChar("0000f00d1212efde1523785fef13d123", "0000feed1212efde1523785fef13d123", "", gatt.CharWrite, 23, 24)
	if err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	if err = periph.WriteCharacteristic(characteristic, []byte{0x01}, false); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	return
}

func (b Microbit) identifyOnPeriphConnected(periph gatt.Peripheral, err error) {
	defer periph.Device().CancelConnection(periph)

	b.Micro.StateChannel <- map[string]bool{
		"connected": true,
	}

	characteristic, err := bluetooth.GetChar("0000f00d1212efde1523785fef13d123", "0000beef1212efde1523785fef13d123", "", gatt.CharWrite, 21, 22)
	if err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	if err = periph.WriteCharacteristic(characteristic, []byte{0x01}, false); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	return
}
