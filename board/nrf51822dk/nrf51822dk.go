package nrf51822dk

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/paypal/gatt"
	"github.com/resin-io/edge-node-manager/micro/nrf51822"
	"github.com/resin-io/edge-node-manager/radio/bluetooth"
)

type Nrf51822dk struct {
	Log   *logrus.Logger
	Micro nrf51822.Nrf51822
}

func (b Nrf51822dk) Update(path string) error {
	b.Log.Info("Starting update")

	if err := b.Micro.ExtractFirmware(path, "nrf51422_xxac_s130.bin", "nrf51422_xxac_s130.dat"); err != nil {
		return err
	}

RetryLoop:
	for i := 1; i <= 3; i++ {
		b.Log.WithFields(logrus.Fields{
			"Number": i,
		}).Info("Update attempt")

		bluetooth.Radio.Handle(
			gatt.PeripheralDiscovered(b.Micro.OnPeriphDiscovered),
			gatt.PeripheralConnected(b.bootloadOnPeriphConnected),
			gatt.PeripheralDisconnected(b.Micro.OnPeriphDisconnected),
		)
		if err := bluetooth.Radio.Init(bluetooth.OnStateChanged); err != nil {
			b.Log.WithFields(logrus.Fields{
				"Error": err,
			}).Error("Update failed")
			return err
		}

		var savedErr error
		var savedRestart bool
		for {
			select {
			case <-time.After(120 * time.Second):
				b.Log.Warn("Timed out")
				continue RetryLoop
			case savedErr = <-b.Micro.ErrChannel:
			case savedRestart = <-b.Micro.RestartChannel:
			case connected := <-b.Micro.ConnectedChannel:
				if connected {
					b.Log.Debug("Connected")
				} else {
					b.Log.Debug("Disconnected")

					if !savedRestart {
						if savedErr == nil {
							b.Log.Info("Finished update")
							return nil
						}

						b.Log.WithFields(logrus.Fields{
							"Error": savedErr,
						}).Error("Update failed")
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
						b.Log.WithFields(logrus.Fields{
							"Error": err,
						}).Error("Update failed")
						return err
					}
				}
			}
		}
	}
	b.Log.Error("Update failed after 3 attempts")
	return fmt.Errorf("Timed out after 3 attempts")
}

func (b Nrf51822dk) Scan(applicationUUID int) (map[string]bool, error) {
	return bluetooth.Scan(strconv.Itoa(applicationUUID), 10)
}

func (b Nrf51822dk) Online() (bool, error) {
	return bluetooth.Online(b.Micro.LocalUUID, 10)
}

func (b Nrf51822dk) Restart() error {
	b.Log.Info("Restarting...")
	return b.Micro.ProcessRequest(b.restartOnPeriphConnected)
}

func (b Nrf51822dk) Identify() error {
	b.Log.Info("Identifying...")
	return b.Micro.ProcessRequest(b.identifyOnPeriphConnected)
}

func (b Nrf51822dk) UpdateConfig(config interface{}) error {
	b.Log.WithFields(logrus.Fields{
		"Config": config,
	}).Info("Updating config...")
	return fmt.Errorf("Update config not implemented")
}

func (b Nrf51822dk) UpdateEnvironment(config interface{}) error {
	b.Log.WithFields(logrus.Fields{
		"Config": config,
	}).Info("Updating environment...")
	return fmt.Errorf("Update environment not implemented")
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
		b.Log.Debug("Bootloader started")
		b.Micro.UpdateOnPeriphConnected(periph, err)
		return
	}

	b.Log.Debug("Bootloader not started")
	b.Log.Debug("Starting bootloader")

	b.Micro.RestartChannel <- true

	if err := b.Micro.EnableCCCD(periph); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	if err := b.Micro.WriteDFUControlPoint(periph, []byte{nrf51822.Start, 0x04}, false); err != nil {
		b.Micro.ErrChannel <- err
		return
	}

	b.Log.Debug("Started bootloader")
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
