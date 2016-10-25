package nrf51822dk

import (
	"fmt"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/micro/nrf51822"
	"github.com/josephroberts/edge-node-manager/radio/bluetooth"
	"github.com/paypal/gatt"
)

type Nrf51822dk struct {
	Micro nrf51822.Nrf51822
}

func (b Nrf51822dk) Update(path string) error {
	if err := b.Micro.ExtractFirmware(path, "nrf51422_xxac_s130.bin", "nrf51422_xxac_s130.dat"); err != nil {
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
		case <-time.After(60 * time.Second):
			return fmt.Errorf("Update timed out, error: %s", savedErr)
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

func (b Nrf51822dk) Scan(applicationUUID int) (map[string]bool, error) {
	return bluetooth.Scan(strconv.Itoa(applicationUUID), 10)
}

func (b Nrf51822dk) Online() (bool, error) {
	return bluetooth.Online(b.Micro.LocalUUID, 10)
}

func (b Nrf51822dk) Restart() error {
	return b.Micro.ProcessRequest(b.restartOnPeriphConnected)
}

func (b Nrf51822dk) Identify() error {
	return b.Micro.ProcessRequest(b.identifyOnPeriphConnected)
}

func (b Nrf51822dk) bootloadOnPeriphConnected(periph gatt.Peripheral, err error) {
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

	if err := b.Micro.EnableCCCD(periph); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	if err := b.Micro.WriteDFUControlPoint(periph, []byte{nrf51822.Start, 0x04}, false); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	log.Debug("Started bootloader")
}

func (b Nrf51822dk) restartOnPeriphConnected(periph gatt.Peripheral, err error) {
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

func (b Nrf51822dk) identifyOnPeriphConnected(periph gatt.Peripheral, err error) {
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
