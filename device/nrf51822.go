package device

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mholt/archiver"
	"github.com/paypal/gatt"

	"github.com/josephroberts/edge-node-manager/firmware"
	"github.com/josephroberts/edge-node-manager/radio/bluetooth"
)

// ALL THESE TYPES WILL BE USED LATER ON @LORENZO

// type operation byte

// /*
//  * Operations defined by device
//  * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
//  */
// const (
// 	START_DFU                   operation = 0x01
// 	INITIALISE_DFU                        = 0x02
// 	RECEIVE_FIRMWARE_IMAGE                = 0x03
// 	VALIDATE_FIRMWARE_IMAGE               = 0x04
// 	ACTIVATE_FIRMWARE_AND_RESET           = 0x05
// 	SYSTEM_RESET                          = 0x06
// 	REPORT_RECEIVED_IMG_SIZE              = 0x07
// 	PKT_RCPT_NOTIF_REQ                    = 0x08
// 	RESPONSE                              = 0x10
// 	PKT_RCPT_NOTIF                        = 0x11
// )

// type response byte

// /*
//  * Responses defined by device
//  * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
//  */
// const (
// 	SUCCESS          response = 0x01
// 	INVALID_STATE             = 0x02
// 	NOT_SUPPORTED             = 0x03
// 	DATA_SIZE                 = 0x04
// 	CRC_ERROR                 = 0x05
// 	OPERATION_FAILED          = 0x06
// )

// type updateType byte

// /*
//  * Update type defined by device
//  * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
//  */
// const (
// 	SOFT_DEVICE            updateType = 0x01
// 	BOOTLOADER                        = 0x02
// 	SOFT_DEVICE_BOOTLOADER            = 0x03
// 	APPLICATION                       = 0x04
// )

// type handle byte

// /*
//  * Handles defined by device
//  * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
//  */
// const (
// 	CONTROL handle = 0x10
// 	PACKET         = 0x0E
// )

// // Defined by firmware running on the device
// const (
// 	PACKET_SIZE     int  = 20
// 	NAME_HANDLE     byte = 0x03
// 	IDENTIFY_HANDLE byte = 0x16
// 	RESTART_HANDLE  byte = 0x18
// )

type Nrf51822 Device

type FOTA struct {
	progress   float32
	startBlock int
	binary     []byte
	data       []byte
	size       int
	state      string
	connected  bool
}

var (
	errChanel   = make(chan error)
	fotaChannel = make(chan FOTA)
	fota        = FOTA{}
)

func (d Nrf51822) String() string {
	return (Device)(d).String()
}

func (d Nrf51822) Update(firmware firmware.Firmware) error {
	log.WithFields(log.Fields{
		"Device":             d,
		"Firmware directory": firmware.Directory,
		"Commit":             firmware.Commit,
	}).Info("Update")

	var err error

	if err = d.extractFirmware(firmware); err != nil {
		return err
	}

	fota.state = "starting"

	bluetooth.Radio.Handle(
		gatt.PeripheralDiscovered(d.onPeriphDiscovered),
		gatt.PeripheralConnected(d.onPeriphConnected),
		gatt.PeripheralDisconnected(d.onPeriphDisconnected),
	)
	if err = bluetooth.Radio.Init(d.onStateChanged); err != nil {
		return err
	}

	for {
		select {
		case err = <-errChanel:
		case fota := <-fotaChannel:
			if fota.connected == true {
				log.Debug("Connected")
			} else {
				log.WithFields(log.Fields{
					"FOTA state": fota.state,
				}).Debug("Disconnected")

				if fota.state == "startBootloader" {
					if err := bluetooth.Radio.Init(d.onStateChanged); err != nil {
						return err
					}
				} else if fota.state == "checkFOTA" {
					return err
				}
			}
		}
	}

	return err
}

func (d Nrf51822) Online() (bool, error) {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Online")
	return bluetooth.Online(d.LocalUUID, 10)
}

func (d Nrf51822) Identify() error {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Identify")
	return nil
}

func (d Nrf51822) Restart() error {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Restart")
	return nil
}

func (d Nrf51822) extractFirmware(firmware firmware.Firmware) error {
	//TODO: maybe worth passing in the final directory instead of directory and commit
	directory := path.Join(firmware.Directory, firmware.Commit)
	var err error

	if err = archiver.Unzip(path.Join(directory, "application.zip"), directory); err != nil {
		return err
	}

	fota.binary, err = ioutil.ReadFile(path.Join(directory, "nrf51422_xxac_s130.bin"))
	if err != nil {
		return err
	}

	fota.data, err = ioutil.ReadFile(path.Join(directory, "nrf51422_xxac_s130.dat"))
	if err != nil {
		return err
	}

	fota.size = len(fota.binary)

	return nil
}

func (d Nrf51822) getName(periph gatt.Peripheral) (string, error) {
	serviceUUID, err := gatt.ParseUUID("1800")
	if err != nil {
		return "", err
	}

	characteristicUUID, err := gatt.ParseUUID("2a00")
	if err != nil {
		return "", err
	}

	service := gatt.NewService(serviceUUID)
	characteristic := gatt.NewCharacteristic(characteristicUUID, service, 0xA, 2, 3)

	byte, err := periph.ReadCharacteristic(characteristic)
	if err != nil {
		return "", err
	}

	return string(byte), nil
}

func (d Nrf51822) startBootloader(periph gatt.Peripheral) error {
	fota.state = "startBootloader"

	name, err := d.getName(periph)
	if err != nil {
		return err
	}

	if name == "DfuTarg" {
		log.Debug("In bootloader mode")
		return nil
	} else {
		log.Debug("Started bootloader mode")
	}

	serviceUUID, err := gatt.ParseUUID("000015301212efde1523785feabcd123")
	if err != nil {
		return err
	}

	characteristicUUID, err := gatt.ParseUUID("000015311212efde1523785feabcd123")
	if err != nil {
		return err
	}

	descriptorUUID, err := gatt.ParseUUID("2902")
	if err != nil {
		return err
	}

	service := gatt.NewService(serviceUUID)
	characteristic := gatt.NewCharacteristic(characteristicUUID, service, 0x18, 15, 16)
	descriptor := gatt.NewDescriptor(descriptorUUID, 17, characteristic)
	characteristic.SetDescriptor(descriptor)

	if err := periph.WriteDescriptor(descriptor, []byte{0x01, 0x00}); err != nil {
		return err
	}

	if err := periph.WriteCharacteristic(characteristic, []byte{0x01, 0x04}, false); err != nil {
		return err
	}

	return nil
}

func (d Nrf51822) checkFOTA(periph gatt.Peripheral) error {
	fota.state = "checkFOTA"

	log.Debug("Checking FOTA") //Why does this get printed twice?

	serviceUUID, err := gatt.ParseUUID("000015301212efde1523785feabcd123")
	if err != nil {
		return err
	}

	characteristicUUID, err := gatt.ParseUUID("000015311212efde1523785feabcd123")
	if err != nil {
		return err
	}

	descriptorUUID, err := gatt.ParseUUID("2902")
	if err != nil {
		return err
	}

	service := gatt.NewService(serviceUUID)
	characteristic := gatt.NewCharacteristic(characteristicUUID, service, 0x18, 15, 16)
	descriptor := gatt.NewDescriptor(descriptorUUID, 17, characteristic)
	characteristic.SetDescriptor(descriptor)

	if err := periph.WriteDescriptor(descriptor, []byte{0x01, 0x00}); err != nil {
		return err
	}

	notifiedChannel := make(chan []byte)
	callback := func(c *gatt.Characteristic, b []byte, err error) {
		notifiedChannel <- b
	}

	if err := periph.SetNotifyValue(characteristic, callback); err != nil {
		return err
	}

	if err := periph.WriteCharacteristic(characteristic, []byte{0x07}, false); err != nil {
		return err
	}

	for {
		select {
		case <-time.After(1 * time.Second):
			return fmt.Errorf("Timed out waiting for notification")
		case response := <-notifiedChannel:
			if (int)(response[0]) != 16 || (int)(response[1]) != 7 || (int)(response[2]) != 1 {
				return fmt.Errorf("Incorrect notification received")
			}

			fota.startBlock = 0
			fota.startBlock += ((int)(response[6]) << 24)
			fota.startBlock += ((int)(response[5]) << 16)
			fota.startBlock += ((int)(response[4]) << 8)
			fota.startBlock += ((int)(response[3]) << 0)
			break
		}
		break
	}

	log.WithFields(log.Fields{
		"Start block": fota.startBlock,
	}).Debug("Checked FOTA")

	return err
}

func (d Nrf51822) initFOTA(periph gatt.Peripheral) error {
	return nil
}

func (d Nrf51822) transferFOTA() error {
	return nil
}

func (d Nrf51822) validateFOTA() error {
	return nil
}

func (d Nrf51822) onStateChanged(radio gatt.Device, state gatt.State) {
	switch state {
	case gatt.StatePoweredOn:
		radio.Scan([]gatt.UUID{}, false)
		return
	default:
		radio.StopScanning()
	}
}

func (d Nrf51822) onPeriphDiscovered(periph gatt.Peripheral, adv *gatt.Advertisement, rssi int) {
	if strings.ToUpper(periph.ID()) != d.LocalUUID {
		return
	}

	periph.Device().StopScanning()
	periph.Device().Connect(periph)
}

func (d Nrf51822) onPeriphConnected(periph gatt.Peripheral, err error) {
	fota.connected = true
	fotaChannel <- fota

	defer periph.Device().CancelConnection(periph)

	if err := periph.SetMTU(500); err != nil {
		errChanel <- err
		return
	}

	if err := d.startBootloader(periph); err != nil {
		errChanel <- err
		return
	}

	// Time delay to allow device to restart in bootloader mode
	time.Sleep(1 * time.Second)

	if err := d.checkFOTA(periph); err != nil {
		errChanel <- err
		return
	}

	if fota.startBlock == 0 {
		if err := d.initFOTA(periph); err != nil { //Do this next
			errChanel <- err
			return
		}
	}

	// Discovery services
	// ss, err := periph.DiscoverServices(nil)
	// if err != nil {
	// 	fmt.Printf("Failed to discover services, err: %s\n", err)
	// 	return
	// }

	// for _, s := range ss {
	// 	msg := "Service: " + s.UUID().String()
	// 	if len(s.Name()) > 0 {
	// 		msg += " (" + s.Name() + ")"
	// 	}
	// 	fmt.Println(msg)

	// 	// Discovery characteristics
	// 	cs, err := periph.DiscoverCharacteristics(nil, s)
	// 	if err != nil {
	// 		fmt.Printf("Failed to discover characteristics, err: %s\n", err)
	// 		continue
	// 	}

	// 	for _, c := range cs {

	// 		msg := "  Characteristic  " + c.UUID().String()
	// 		if len(c.Name()) > 0 {
	// 			msg += " (" + c.Name() + ")"
	// 		}
	// 		msg += "\n    properties    " + c.Properties().String()
	// 		fmt.Println(msg)

	// 		fmt.Println("H:", c.Handle())
	// 		fmt.Println("VH: ", c.VHandle())

	// 		// Read the characteristic, if possible.
	// 		if (c.Properties() & gatt.CharRead) != 0 {
	// 			b, err := periph.ReadCharacteristic(c)
	// 			if err != nil {
	// 				fmt.Printf("Failed to read characteristic, err: %s\n", err)
	// 				continue
	// 			}
	// 			fmt.Printf("    value         %x | %q\n", b, b)
	// 		}

	// 		// Discovery descriptors
	// 		ds, err := periph.DiscoverDescriptors(nil, c)
	// 		if err != nil {
	// 			fmt.Printf("Failed to discover descriptors, err: %s\n", err)
	// 			continue
	// 		}

	// 		for _, d := range ds {
	// 			msg := "  Descriptor      " + d.UUID().String()
	// 			if len(d.Name()) > 0 {
	// 				msg += " (" + d.Name() + ")"
	// 			}
	// 			fmt.Println(msg)

	// 			fmt.Println("H:", d.Handle())
	// 			// Read descriptor (could fail, if it's not readable)
	// 			b, err := periph.ReadDescriptor(d)
	// 			if err != nil {
	// 				fmt.Printf("Failed to read descriptor, err: %s\n", err)
	// 				continue
	// 			}
	// 			fmt.Printf("    value         %x | %q\n", b, b)

	// 		}

	// 		// Subscribe the characteristic, if possible.
	// 		if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
	// 			f := func(c *gatt.Characteristic, b []byte, err error) {
	// 				fmt.Printf("notified: % X | %q\n", b, b)
	// 			}
	// 			if err := periph.SetNotifyValue(c, f); err != nil {
	// 				fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
	// 				continue
	// 			}
	// 		}

	// 	}
	// 	fmt.Println()
	// }
}

func (d Nrf51822) onPeriphDisconnected(periph gatt.Peripheral, err error) {
	fota.connected = false
	fotaChannel <- fota
}
