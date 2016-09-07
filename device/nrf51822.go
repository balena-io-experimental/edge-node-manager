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

//TODOS
//Identify and restart

/*
 * Please see the links below for an explanation of FOTA on an NRF51822 board
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleprofile.html
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
 */

/*
 * Output from the print function
 *
 * Service: 1800 (Generic Access)
 *   Characteristic  2a00 (Device Name)
 *     properties    read write
 * H:  2  VH:  3
 *     value         726573696e | "resin"
 *   Characteristic  2a01 (Appearance)
 *     properties    read
 * H:  4  VH:  5
 *     value         0000 | "\x00\x00"
 *   Characteristic  2a04 (Peripheral Preferred Connection Parameters)
 *     properties    read
 * H:  6  VH:  7
 *     value         0600200000009001 | "\x06\x00 \x00\x00\x00\x90\x01"
 *
 * Service: 1801 (Generic Attribute)
 *   Characteristic  2a05 (Service Changed)
 *     properties    indicate
 * H:  9  VH:  10
 *   Descriptor      2902 (Client Characteristic Configuration)
 * H:  9
 *     value         0000 | "\x00\x00"
 *
 * Service: 000015301212efde1523785feabcd123
 *   Characteristic  000015321212efde1523785feabcd123
 *     properties    writeWithoutResponse
 * H:  13  VH:  14
 *   Characteristic  000015311212efde1523785feabcd123
 *     properties    write notify
 * H:  15  VH:  16
 *   Descriptor      2902 (Client Characteristic Configuration)
 * H:  15
 *     value         0000 | "\x00\x00"
 *   Characteristic  000015341212efde1523785feabcd123
 *     properties    read
 * H:  18  VH:  19
 *     value         0100 | "\x01\x00"
 *
 * Service: 0000f00d1212efde1523785fef13d123
 *   Characteristic  0000beef1212efde1523785fef13d123
 *     properties    write
 * H:  21  VH:  22
 *   Characteristic  0000feed1212efde1523785fef13d123
 *     properties    write
 * H:  23  VH:  24
 */

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

const (
	SUCCESS          byte = 0x01
	INVALID_STATE         = 0x02
	NOT_SUPPORTED         = 0x03
	DATA_SIZE             = 0x04
	CRC_ERROR             = 0x05
	OPERATION_FAILED      = 0x06
)

const (
	START_DATA  byte = 0x00
	FINISH_DATA      = 0x01
)

const (
	SOFT_DEVICE            byte = 0x01
	BOOTLOADER                  = 0x02
	SOFT_DEVICE_BOOTLOADER      = 0x03
	APPLICATION                 = 0x04
)

type Nrf51822 Device

type FOTA struct {
	progress     float32
	currentBlock int
	binary       []byte
	data         []byte
	size         int
	restart      bool
	connected    bool
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
		"Firmware directory": firmware.Dir,
		"Commit":             firmware.Commit,
	}).Info("Update")

	var err error

	if err = d.extractFirmware(firmware); err != nil {
		return err
	}

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
			if fota.connected {
				log.Debug("Connected")
			} else {
				log.Debug("Disconnected")

				if fota.restart {
					/*
					 * The device is expected to restart after being placed into bootloader mode
					 * so it is necessary to re-connect afterwards
					 */
					if err := bluetooth.Radio.Init(d.onStateChanged); err != nil {
						return err
					}
				}
				return err
			}
		}
	}
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

	// serviceUUID, err := gatt.ParseUUID("0000f00d1212efde1523785fef13d123")
	// if err != nil {
	// 	return err
	// }

	// characteristicUUID, err := gatt.ParseUUID("0000beef1212efde1523785fef13d123")
	// if err != nil {
	// 	return err
	// }

	// service := gatt.NewService(serviceUUID)
	// characteristic := gatt.NewCharacteristic(characteristicUUID, service, gatt.CharWrite, 21, 22)

	// err := periph.WriteCharacteristic(characteristic, []byte{0x01}, false)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (d Nrf51822) Restart() error {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Restart")

	// serviceUUID, err := gatt.ParseUUID("0000f00d1212efde1523785fef13d123")
	// if err != nil {
	// 	return err
	// }

	// characteristicUUID, err := gatt.ParseUUID("0000feed1212efde1523785fef13d123")
	// if err != nil {
	// 	return err
	// }

	// service := gatt.NewService(serviceUUID)
	// characteristic := gatt.NewCharacteristic(characteristicUUID, service, gatt.CharWrite, 23, 24)

	// err := periph.WriteCharacteristic(characteristic, []byte{0x01}, false)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (d Nrf51822) extractFirmware(firmware firmware.Firmware) error {
	var err error

	if err = archiver.Unzip(path.Join(firmware.Dir, "application.zip"), firmware.Dir); err != nil {
		return err
	}

	fota.binary, err = ioutil.ReadFile(path.Join(firmware.Dir, "nrf51422_xxac_s130.bin"))
	if err != nil {
		return err
	}

	fota.data, err = ioutil.ReadFile(path.Join(firmware.Dir, "nrf51422_xxac_s130.dat"))
	if err != nil {
		return err
	}

	fota.size = len(fota.binary)

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
	defer periph.Device().CancelConnection(periph)

	fota.connected = true
	fotaChannel <- fota

	if log.GetLevel() == log.DebugLevel {
		if err := d.print(periph); err != nil {
			errChanel <- err
			return
		}
	}

	if err := periph.SetMTU(500); err != nil {
		errChanel <- err
		return
	}

	if err := d.startBootloader(periph); err != nil {
		errChanel <- err
		return
	}

	// Time delay to allow device restart after being placed into bootloader mode
	time.Sleep(1 * time.Second)

	if err := d.checkFOTA(periph); err != nil {
		errChanel <- err
		return
	}

	if fota.currentBlock == 0 {
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
	log.Debug("Starting bootloader mode")

	name, err := d.getName(periph)
	if err != nil {
		return err
	}

	if name == "DfuTarg" {
		log.Debug("In bootloader mode")
		fota.restart = false
		return nil
	}

	fota.restart = true

	if err := d.enableCCCD(periph); err != nil {
		return err
	}

	if err := d.writeDFUControlPoint(periph, []byte{START_DFU, APPLICATION}, false); err != nil {
		return err
	}

	log.Debug("Started bootloader mode")

	return nil
}

func (d Nrf51822) checkFOTA(periph gatt.Peripheral) error {
	// TODO: figure out why we get this twice - something weird going on with the control flow
	log.Debug("Checking FOTA")

	if err := d.enableCCCD(periph); err != nil {
		return err
	}

	resp, err := d.notifyDFUControlPoint(periph, []byte{REPORT_RECEIVED_IMG_SIZE})
	if err != nil {
		return err
	}

	if resp[0] != RESPONSE || resp[1] != REPORT_RECEIVED_IMG_SIZE || resp[2] != SUCCESS {
		return fmt.Errorf("Incorrect notification received")
	}

	fota.currentBlock, err = d.parseResp(resp[3:])
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Start block": fota.currentBlock,
	}).Debug("Checked FOTA")

	return err
}

func (d Nrf51822) initFOTA(periph gatt.Peripheral) error {
	log.Debug("Initialising FOTA")

	if err := d.enableCCCD(periph); err != nil {
		return err
	}

	if err := d.writeDFUControlPoint(periph, []byte{START_DFU, APPLICATION}, false); err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	// Pad the buffer with 8 zeroed bytes
	if _, err := buf.Write(make([]byte, 8)); err != nil {
		return err
	}

	if err := binary.Write(buf, binary.LittleEndian, (int32)(fota.size)); err != nil {
		return err
	}

	resp, err := d.notifyDFUPacket(periph, buf.Bytes())
	if err != nil {
		return err
	}

	if resp[0] != RESPONSE || resp[1] != START_DFU || resp[2] != SUCCESS {
		return fmt.Errorf("Incorrect notification received")
	}

	if err := d.writeDFUControlPoint(periph, []byte{INITIALISE_DFU, START_DATA}, false); err != nil {
		return err
	}

	if err := d.writeDFUPacket(periph, fota.data, false); err != nil {
		return err
	}

	resp, err = d.notifyDFUControlPoint(periph, []byte{INITIALISE_DFU, FINISH_DATA})
	if err != nil {
		return err
	}

	if resp[0] != RESPONSE || resp[1] != INITIALISE_DFU || resp[2] != SUCCESS {
		return fmt.Errorf("Incorrect notification received")
	}

	if err := d.writeDFUControlPoint(periph, []byte{PKT_RCPT_NOTIF_REQ, 0x64, 0x00}, false); err != nil {
		return err
	}

	if err := d.writeDFUControlPoint(periph, []byte{RECEIVE_FIRMWARE_IMAGE}, false); err != nil {
		return err
	}

	log.Debug("Initialised FOTA")

	return nil
}

func (d Nrf51822) transferFOTA(periph gatt.Peripheral) error {
	blockCounter := 1
	blockSize := 20
	if fota.currentBlock != 0 {
		/*
		 * Set block counter to the current block, this is used to resume FOTA if
		 * the previous FOTA was cancelled/failed mid way though
		 */
		blockCounter += (fota.currentBlock / blockSize)
	}

	fota.progress = ((float32)(fota.currentBlock) / (float32)(fota.size)) * 100.0

	log.WithFields(log.Fields{
		"Block counter": blockCounter,
		"Progress %":    fota.progress,
	}).Debug("Transferring FOTA")

	notifyChannel, err := d.initNotify(periph)
	if err != nil {
		return err
	}

	for i := fota.currentBlock; i < fota.size; i += blockSize {
		// Extract the block from the binary
		sliceIndex := i + blockSize
		if sliceIndex > fota.size {
			// Limit the slice to fota.Size to avoid extra zeros being tagged on the end of the block
			sliceIndex = fota.size
		}
		block := fota.binary[i:sliceIndex]

		if err := d.writeDFUPacket(periph, block, false); err != nil {
			return err
		}

		if (blockCounter % 100) == 0 {
			resp, err := d.timeoutNotify(notifyChannel)
			if err != nil {
				return err
			}

			if resp[0] != PKT_RCPT_NOTIF {
				return fmt.Errorf("Incorrect notification received")
			}

			currentBlock, err := d.parseResp(resp[1:])
			if err != nil {
				return err
			}

			if (i + blockSize) != currentBlock {
				return fmt.Errorf("FOTA transer out of sync")
			}

			fota.progress = ((float32)(currentBlock) / (float32)(fota.size)) * 100.0

			log.WithFields(log.Fields{
				"Block counter": blockCounter,
				"Progress %":    fota.progress,
			}).Debug("Transferring FOTA")
		}

		blockCounter++
	}

	fota.progress = 100

	resp, err := d.timeoutNotify(notifyChannel)
	if err != nil {
		return err
	}

	if resp[0] != RESPONSE || resp[1] != RECEIVE_FIRMWARE_IMAGE || resp[2] != SUCCESS {
		return fmt.Errorf("Incorrect notification received")
	}

	log.Debug("Transferred FOTA")

	return nil
}

func (d Nrf51822) validateFOTA(periph gatt.Peripheral) error {
	log.Debug("Validating FOTA")

	if err := d.checkFOTA(periph); err != nil {
		return err
	}

	if fota.currentBlock != fota.size {
		return fmt.Errorf("Bytes received does not match binary size")
	}

	resp, err := d.notifyDFUControlPoint(periph, []byte{VALIDATE_FIRMWARE_IMAGE})
	if err != nil {
		return err
	}

	if resp[0] != RESPONSE || resp[1] != VALIDATE_FIRMWARE_IMAGE || resp[2] != SUCCESS {
		return fmt.Errorf("Incorrect notification received")
	}

	log.Debug("Validated FOTA")

	return nil
}

func (d Nrf51822) finaliseFOTA(periph gatt.Peripheral) error {
	log.Debug("Finalising FOTA")

	if err := d.writeDFUControlPoint(periph, []byte{ACTIVATE_FIRMWARE_AND_RESET}, false); err != nil {
		return err
	}

	log.Debug("Finalised FOTA")

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

func (d Nrf51822) writeDFUControlPoint(periph gatt.Peripheral, value []byte, noRsp bool) error {
	characteristic, err := d.getChar("000015311212efde1523785feabcd123", gatt.CharWrite+gatt.CharNotify, 15, 16)
	if err != nil {
		return err
	}

	return periph.WriteCharacteristic(characteristic, value, noRsp)
}

func (d Nrf51822) writeDFUPacket(periph gatt.Peripheral, value []byte, noRsp bool) error {
	characteristic, err := d.getChar("000015321212efde1523785feabcd123", gatt.CharWriteNR, 13, 14)
	if err != nil {
		return err
	}

	return periph.WriteCharacteristic(characteristic, value, noRsp)
}

func (d Nrf51822) notifyDFUControlPoint(periph gatt.Peripheral, value []byte) ([]byte, error) {
	notifyChannel, err := d.initNotify(periph)
	if err != nil {
		return nil, err
	}

	if err := d.writeDFUControlPoint(periph, value, false); err != nil {
		return nil, err
	}

	return d.timeoutNotify(notifyChannel)

}

func (d Nrf51822) notifyDFUPacket(periph gatt.Peripheral, value []byte) ([]byte, error) {
	notifyChannel, err := d.initNotify(periph)
	if err != nil {
		return nil, err
	}

	if err := d.writeDFUPacket(periph, value, false); err != nil {
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

	err = periph.SetNotifyValue(characteristic, callback)

	return notifyChannel, err
}

func (d Nrf51822) timeoutNotify(notifyChannel chan []byte) ([]byte, error) {
	for {
		select {
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("Timed out waiting for notification")
		case resp := <-notifyChannel:
			log.WithFields(log.Fields{
				"[0]": resp[0],
				"[1]": resp[1],
				"[2]": resp[2],
			}).Debug("notification")
			return resp, nil
		}
	}
}

func (d Nrf51822) parseResp(resp []byte) (int, error) {
	var result int32
	buf := bytes.NewReader(resp)
	if err := binary.Read(buf, binary.LittleEndian, &result); err != nil {
		return 0, err
	}
	return (int)(result), nil
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

			fmt.Println("H: ", c.Handle(), " VH: ", c.VHandle())

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

				fmt.Println("H: ", c.Handle())

				b, err := periph.ReadDescriptor(d)
				if err != nil {
					fmt.Printf("Failed to read descriptor, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}
		}
		fmt.Println()
	}

	return nil
}
