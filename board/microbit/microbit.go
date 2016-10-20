package microbit

import (
	"fmt"

	"github.com/josephroberts/edge-node-manager/micro/nrf51822"
)

type MicroBit struct {
	Micro nrf51822.Nrf51822
}

func (d MicroBit) Update(path string) error {
	fmt.Println("MicroBit update")
	return nil
}

func (d MicroBit) Scan() (map[string]bool, error) {
	fmt.Println("MicroBit scan")
	return nil, nil
}

func (d MicroBit) Online() (bool, error) {
	fmt.Println("MicroBit online")
	return false, nil
}

func (d MicroBit) Restart() error {
	fmt.Println("MicroBit restart")
	return nil
}

func (d MicroBit) Identify() error {
	fmt.Println("MicroBit identify")
	return nil
}

// // Update updates the device
// func (d MicroBit) Update(path string) error {
// 	fmt.Println("MicroBit update")
// 	return nil
// }

// 	if err := d.Nrf51822.ExtractFirmware(path, "micro-bit.bin", "micro-bit.dat"); err != nil {
// 		return err
// 	}

// 	bluetooth.Radio.Handle(
// 		gatt.PeripheralDiscovered(d.Nrf51822.OnPeriphDiscovered),
// 		gatt.PeripheralConnected(d.bootloadOnPeriphConnected),
// 		gatt.PeripheralDisconnected(d.Nrf51822.OnPeriphDisconnected),
// 	)
// 	if err := bluetooth.Radio.Init(bluetooth.OnStateChanged); err != nil {
// 		return err
// 	}

// 	var savedErr error
// 	for {
// 		select {
// 		case savedErr = <-d.Nrf51822.ErrChannel:
// 		case state := <-d.Nrf51822.FotaChannel:
// 			if state.Connected {
// 				log.Debug("Connected")
// 			} else {
// 				log.Debug("Disconnected")

// 				if !state.Restart {
// 					return savedErr
// 				}

// 				bluetooth.Radio.Handle(
// 					gatt.PeripheralDiscovered(d.Nrf51822.OnPeriphDiscovered),
// 					gatt.PeripheralConnected(d.Nrf51822.UpdateOnPeriphConnected),
// 					gatt.PeripheralDisconnected(d.Nrf51822.OnPeriphDisconnected),
// 				)
// 				if err := bluetooth.Radio.Init(bluetooth.OnStateChanged); err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}
// }

// // Scan checks which devices are online
// func (d MicroBit) Scan() (map[string]bool, error) {
// 	id := "BBC micro:bit [" + strconv.Itoa(d.ApplicationUUID) + "]"
// 	return bluetooth.Scan(id, 10)
// }

// // Online checks whether the device is online
// func (d MicroBit) Online() (bool, error) {
// 	id := "BBC micro:bit [" + strconv.Itoa(d.ApplicationUUID) + "]"
// 	return bluetooth.Online(id, 10)
// }

// // Restart restarts the device
// func (d MicroBit) Restart() error {
// 	return d.Nrf51822.ProcessRequest(d.restartOnPeriphConnected)
// }

// // Identify flashes LEDs' on the device
// func (d MicroBit) Identify() error {
// 	return d.Nrf51822.ProcessRequest(d.identifyOnPeriphConnected)
// }

// func (d MicroBit) bootloadOnPeriphConnected(periph gatt.Peripheral, err error) {
// 	defer periph.Device().CancelConnection(periph)

// 	d.Nrf51822.Fota.Connected = true
// 	d.Nrf51822.FotaChannel <- d.Nrf51822.Fota

// 	if err := periph.SetMTU(500); err != nil {
// 		d.Nrf51822.ErrChannel <- err
// 		return
// 	}

// 	if err := d.startBootloader(periph); err != nil {
// 		d.Nrf51822.ErrChannel <- err
// 		return
// 	}

// 	// Time delay to allow device restart after being placed into bootloader mode
// 	time.Sleep(1 * time.Second)
// }

// func (d MicroBit) identifyOnPeriphConnected(periph gatt.Peripheral, err error) {
// 	defer periph.Device().CancelConnection(periph)

// 	d.Nrf51822.Fota.Connected = true
// 	d.Nrf51822.FotaChannel <- d.Nrf51822.Fota

// 	characteristic, err := bluetooth.GetChar("0000f00d1212efde1523785fef13d123", "0000beef1212efde1523785fef13d123", "", gatt.CharWrite, 21, 22)
// 	if err != nil {
// 		d.Nrf51822.ErrChannel <- err
// 		return
// 	}

// 	if err = periph.WriteCharacteristic(characteristic, []byte{0x01}, false); err != nil {
// 		d.Nrf51822.ErrChannel <- err
// 		return
// 	}

// 	return
// }

// func (d MicroBit) restartOnPeriphConnected(periph gatt.Peripheral, err error) {
// 	defer periph.Device().CancelConnection(periph)

// 	d.Nrf51822.Fota.Connected = true
// 	d.Nrf51822.FotaChannel <- d.Nrf51822.Fota

// 	characteristic, err := bluetooth.GetChar("0000f00d1212efde1523785fef13d123", "0000feed1212efde1523785fef13d123", "", gatt.CharWrite, 23, 24)
// 	if err != nil {
// 		d.Nrf51822.ErrChannel <- err
// 		return
// 	}

// 	if err = periph.WriteCharacteristic(characteristic, []byte{0x01}, false); err != nil {
// 		d.Nrf51822.ErrChannel <- err
// 		return
// 	}

// 	return
// }

// func (d MicroBit) startBootloader(periph gatt.Peripheral) error {
// 	log.Debug("Starting bootloader mode")

// 	name, err := bluetooth.GetName(periph)
// 	if err != nil {
// 		return err
// 	}

// 	// The device name is used to check whether the device is in bootloader mode
// 	if name == "DfuTarg" {
// 		log.Debug("In bootloader mode")
// 		d.Nrf51822.Fota.Restart = false
// 		return nil
// 	}

// 	d.Nrf51822.Fota.Restart = true

// 	if err = d.writeEnableDFUControlPoint(periph, []byte{nrf51822.Start}, false); err != nil {
// 		return err
// 	}

// 	log.Debug("Started bootloader mode")

// 	return nil
// }

// func (d MicroBit) writeEnableDFUControlPoint(periph gatt.Peripheral, value []byte, noRsp bool) error {
// 	characteristic, err := bluetooth.GetChar("e95d93b0251d470aa062fa1922dfa9a8", "e95d93b1251d470aa062fa1922dfa9a8", "2902", gatt.CharRead+gatt.CharWrite, 13, 14)
// 	if err != nil {
// 		return err
// 	}

// 	return periph.WriteCharacteristic(characteristic, value, noRsp)
// }
