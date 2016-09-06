package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mholt/archiver"
	"github.com/paypal/gatt"

	"github.com/josephroberts/edge-node-manager/firmware"
	"github.com/josephroberts/edge-node-manager/radio/bluetooth"
)

/*
 * Operations defined by device
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
 */
// TODO: Make these generic for operation and procedure
const (
	START_DFU                   byte = 0x01
	INITIALISE_DFU                   = 0x02
	RECEIVE_FIRMWARE_IMAGE           = 0x03
	VALIDATE_FIRMWARE_IMAGE          = 0x04
	ACTIVATE_FIRMWARE_AND_RESET      = 0x05
	SYSTEM_RESET                     = 0x06
	REPORT_RECEIVED_IMG_SIZE         = 0x07
	PKT_RCPT_NOTIF_REQ               = 0x08
	RESPONSE                         = 0x10
	PKT_RCPT_NOTIF                   = 0x11
)

/*
 * Responses defined by device
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
 */
const (
	SUCCESS          byte = 0x01
	INVALID_STATE         = 0x02
	NOT_SUPPORTED         = 0x03
	DATA_SIZE             = 0x04
	CRC_ERROR             = 0x05
	OPERATION_FAILED      = 0x06
)

/*
 * Update type defined by device
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
 */
//TODO: the comment above
const (
	START_DATA  byte = 0x00
	FINISH_DATA      = 0x01
)

/*
 * Update type defined by device
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
 */
const (
	SOFT_DEVICE            byte = 0x01
	BOOTLOADER                  = 0x02
	SOFT_DEVICE_BOOTLOADER      = 0x03
	APPLICATION                 = 0x04
)

type Nrf51822 Device

type FOTA struct {
	progress   float32
	startBlock int
	binary     []byte
	data       []byte
	size       int32
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

	fota.state = "extractingFOTA"

	if err = d.extractFirmware(firmware); err != nil {
		return err
	}

	fota.state = "startingFOTA"

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
				} else if fota.state == "initFOTA" {
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

	fota.size = (int32)(len(fota.binary))

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
	if periph.ID() != d.LocalUUID {
		return
	}

	periph.Device().StopScanning()
	periph.Device().Connect(periph)
}

func (d Nrf51822) onPeriphConnected(periph gatt.Peripheral, err error) {
	fota.connected = true
	fotaChannel <- fota

	//d.print(periph)

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
		if err := d.initFOTA(periph); err != nil {
			errChanel <- err
			return
		}
	}

	if err := d.transferFOTA(periph); err != nil {
		errChanel <- err
		return
	}

	if err := d.validateFOTA(periph); err != nil {
		errChanel <- err
		return
	}

	if err := d.finaliseFOTA(periph); err != nil {
		errChanel <- err
		return
	}
}

func (d Nrf51822) onPeriphDisconnected(periph gatt.Peripheral, err error) {
	fota.connected = false
	fotaChannel <- fota
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
	characteristic := gatt.NewCharacteristic(characteristicUUID, service, gatt.CharRead+gatt.CharWrite, 2, 3)

	byte, err := periph.ReadCharacteristic(characteristic)
	if err != nil {
		return "", err
	}

	return string(byte), nil
}

func (d Nrf51822) startBootloader(periph gatt.Peripheral) error {
	fota.state = "startBootloader"

	log.Debug("Starting bootloader mode")

	name, err := d.getName(periph)
	if err != nil {
		return err
	}

	if name == "DfuTarg" {
		log.Debug("In bootloader mode")
		return nil
	}

	if err := d.enableCCCD(periph); err != nil {
		return err
	}

	if err := d.writeDFUControlPoint(periph, []byte{START_DFU, APPLICATION}); err != nil {
		return err
	}

	log.Debug("Started bootloader mode")

	return nil
}

func (d Nrf51822) checkFOTA(periph gatt.Peripheral) error {
	fota.state = "checkFOTA"

	// TODO: figure out why we get this twice - something wierd going on with the control flow
	log.Debug("Checking FOTA")

	if err := d.enableCCCD(periph); err != nil {
		return err
	}

	response, err := d.notifyDFUControlPoint(periph, []byte{REPORT_RECEIVED_IMG_SIZE})
	if err != nil {
		return err
	}

	if response[0] != RESPONSE || response[1] != REPORT_RECEIVED_IMG_SIZE || response[2] != SUCCESS {
		return fmt.Errorf("Incorrect notification received")
	}

	// TODO: encode package
	// Something like ...
	// result := (uint32)(0)
	// b := []byte{0, 0, 0, 0, 0, 0, 0}
	// buf := bytes.NewReader(b[3:])
	// err := binary.Read(buf, binary.LittleEndian, &result)
	// if err != nil {
	// 	fmt.Println("binary.Read failed:", err)
	// }
	// fmt.Print(result)
	fota.startBlock = 0
	fota.startBlock += ((int)(response[3]) << 0)
	fota.startBlock += ((int)(response[4]) << 8)
	fota.startBlock += ((int)(response[5]) << 16)
	fota.startBlock += ((int)(response[6]) << 24)

	log.WithFields(log.Fields{
		"Start block": fota.startBlock,
	}).Debug("Checked FOTA")

	return err
}

func (d Nrf51822) initFOTA(periph gatt.Peripheral) error {
	fota.state = "initFOTA"

	log.Debug("Initialising FOTA")

	if err := d.enableCCCD(periph); err != nil {
		return err
	}

	if err := d.writeDFUControlPoint(periph, []byte{START_DFU, APPLICATION}); err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	// Pad the buffer with 8 zeroed bytes
	buf.Write(make([]byte, 8))
	if err := binary.Write(buf, binary.LittleEndian, fota.size); err != nil {
		return err
	}

	response, err := d.notifyDFUPacket(periph, buf.Bytes())
	if err != nil {
		return err
	}

	if response[0] != RESPONSE || response[1] != START_DFU || response[2] != SUCCESS {
		return fmt.Errorf("Incorrect notification received")
	}

	if err := d.writeDFUControlPoint(periph, []byte{INITIALISE_DFU, START_DATA}); err != nil {
		return err
	}

	if err := d.writeDFUPacket(periph, fota.data); err != nil {
		return err
	}

	response, err = d.notifyDFUControlPoint(periph, []byte{INITIALISE_DFU, FINISH_DATA})
	if err != nil {
		return err
	}

	if response[0] != RESPONSE || response[1] != INITIALISE_DFU || response[2] != SUCCESS {
		return fmt.Errorf("Incorrect notification received")
	}

	if err := d.writeDFUControlPoint(periph, []byte{PKT_RCPT_NOTIF_REQ, 0x64, 0x00}); err != nil {
		return err
	}

	if err := d.writeDFUControlPoint(periph, []byte{RECEIVE_FIRMWARE_IMAGE}); err != nil {
		return err
	}

	return nil
}

func (d Nrf51822) transferFOTA(periph gatt.Peripheral) error {
	return nil
}

func (d Nrf51822) validateFOTA(periph gatt.Peripheral) error {
	return nil
}

func (d Nrf51822) finaliseFOTA(periph gatt.Peripheral) error {
	return nil
}

func (d Nrf51822) getChar(UUID string, props gatt.Property, h, vh uint16) (*gatt.Characteristic, error) {
	serviceUUID, err := gatt.ParseUUID("000015301212efde1523785feabcd123")
	if err != nil {
		return &gatt.Characteristic{}, err
	}

	characteristicUUID, err := gatt.ParseUUID(UUID)
	if err != nil {
		return &gatt.Characteristic{}, err
	}

	descriptorUUID, err := gatt.ParseUUID("2902")
	if err != nil {
		return &gatt.Characteristic{}, err
	}

	service := gatt.NewService(serviceUUID)
	characteristic := gatt.NewCharacteristic(characteristicUUID, service, props, h, vh)
	descriptor := gatt.NewDescriptor(descriptorUUID, 17, characteristic)
	characteristic.SetDescriptor(descriptor)

	return characteristic, nil
}

func (d Nrf51822) enableCCCD(periph gatt.Peripheral) error {
	characteristic, err := d.getChar("000015311212efde1523785feabcd123", gatt.CharWrite+gatt.CharNotify, 15, 16)
	if err != nil {
		return err
	}

	return periph.WriteDescriptor(characteristic.Descriptor(), []byte{START_DFU, 0x00})
}

func (d Nrf51822) writeDFUControlPoint(periph gatt.Peripheral, value []byte) error {
	characteristic, err := d.getChar("000015311212efde1523785feabcd123", gatt.CharWrite+gatt.CharNotify, 15, 16)
	if err != nil {
		return err
	}

	return periph.WriteCharacteristic(characteristic, value, false)
}

func (d Nrf51822) writeDFUPacket(periph gatt.Peripheral, value []byte) error {
	characteristic, err := d.getChar("000015321212efde1523785feabcd123", gatt.CharWriteNR, 13, 14)
	if err != nil {
		return err
	}

	return periph.WriteCharacteristic(characteristic, value, false)
}

func (d Nrf51822) notifyDFUControlPoint(periph gatt.Peripheral, value []byte) ([]byte, error) {
	notifyChannel, err := d.initNotify(periph)
	if err != nil {
		return nil, err
	}

	if err := d.writeDFUControlPoint(periph, value); err != nil {
		return nil, err
	}

	return d.timeoutNotify(notifyChannel)

}

func (d Nrf51822) notifyDFUPacket(periph gatt.Peripheral, value []byte) ([]byte, error) {
	notifyChannel, err := d.initNotify(periph)
	if err != nil {
		return nil, err
	}

	if err := d.writeDFUPacket(periph, value); err != nil {
		return nil, err
	}

	return d.timeoutNotify(notifyChannel)
}

func (d Nrf51822) initNotify(periph gatt.Peripheral) (chan []byte, error) {
	notifyChannel := make(chan []byte)

	characteristic, err := d.getChar("000015311212efde1523785feabcd123", gatt.CharWrite+gatt.CharNotify, 15, 16)
	if err != nil {
		return notifyChannel, err
	}

	callback := func(c *gatt.Characteristic, b []byte, err error) {
		notifyChannel <- b
	}

	if err := periph.SetNotifyValue(characteristic, callback); err != nil {
		return notifyChannel, err
	}

	return notifyChannel, nil
}

func (d Nrf51822) timeoutNotify(notifyChannel chan []byte) ([]byte, error) {
	for {
		select {
		case <-time.After(10 * time.Second): // TODO: set to 1 sec
			return nil, fmt.Errorf("Timed out waiting for notification")
		case response := <-notifyChannel:
			log.WithFields(log.Fields{
				"[0]": response[0],
				"[1]": response[1],
				"[2]": response[2],
			}).Debug("notification")
			return response, nil
		}
	}
}

func (d Nrf51822) print(periph gatt.Peripheral) error {
	ss, err := periph.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return err
	}

	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		cs, err := periph.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			msg := "  Characteristic  " + c.UUID().String()
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
			fmt.Println(msg)

			fmt.Println("H:", c.Handle())
			fmt.Println("VH: ", c.VHandle())

			if (c.Properties() & gatt.CharRead) != 0 {
				b, err := periph.ReadCharacteristic(c)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			ds, err := periph.DiscoverDescriptors(nil, c)
			if err != nil {
				fmt.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			for _, d := range ds {
				msg := "  Descriptor      " + d.UUID().String()
				if len(d.Name()) > 0 {
					msg += " (" + d.Name() + ")"
				}
				fmt.Println(msg)

				fmt.Println("H:", d.Handle())

				b, err := periph.ReadDescriptor(d)
				if err != nil {
					fmt.Printf("Failed to read descriptor, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Subscribe the characteristic, if possible.
			if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
				f := func(c *gatt.Characteristic, b []byte, err error) {
					fmt.Printf("notified: % X | %q\n", b, b)
				}
				if err := periph.SetNotifyValue(c, f); err != nil {
					fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
					continue
				}
			}

		}
		fmt.Println()
	}

	return nil
}
