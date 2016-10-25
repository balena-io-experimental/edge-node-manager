package microbit

import (
	"fmt"
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
	var savedRestart bool
	for {
		select {
		case <-time.After(120 * time.Second):
			return fmt.Errorf("Update timed out")
		case savedErr = <-b.Micro.ErrChannel:
		case savedRestart = <-b.Micro.RestartChannel:
		case connected := <-b.Micro.ConnectedChannel:
			if connected {
				log.Debug("Connected")
			} else {
				log.Debug("Disconnected")

				if !savedRestart {
					return savedErr
				}

				savedRestart = false

				// Time delay to allow device to restart after being placed into bootloader mode
				time.Sleep(1 * time.Second)

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

	b.Micro.ConnectedChannel <- true

	if err := periph.SetMTU(500); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	name, err := bluetooth.GetName(periph)
	if err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	if name == "DfuTarg" {
		log.Debug("Bootloader started")
		b.Micro.UpdateOnPeriphConnected(periph, err)
		return
	}

	log.Debug("Bootloader not started")
	log.Debug("Starting bootloader")

	b.Micro.RestartChannel <- true

	characteristic, err := bluetooth.GetChar("e95d93b0251d470aa062fa1922dfa9a8", "e95d93b1251d470aa062fa1922dfa9a8", "2902", gatt.CharRead+gatt.CharWrite, 13, 14)
	if err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	if err = periph.WriteCharacteristic(characteristic, []byte{nrf51822.Start}, false); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	log.Debug("Started bootloader")
}

func (b Microbit) restartOnPeriphConnected(periph gatt.Peripheral, err error) {
	defer periph.Device().CancelConnection(periph)

	b.Micro.ConnectedChannel <- true

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

	b.Micro.ConnectedChannel <- true

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
