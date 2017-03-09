package nrf51822

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/currantlabs/ble"
	"github.com/mholt/archiver"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/radio/bluetooth"
)

// Firmware-over-the-air update info
// https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleprofile.html
// https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state

const (
	Success            byte = 0x01
	Start                   = 0x01
	Initialise              = 0x02
	Receive                 = 0x03
	Validate                = 0x04
	Activate                = 0x05
	Restart                 = 0x06
	ReceivedSize            = 0x07
	RequestBlockRecipt      = 0x08
	Response                = 0x10
	BlockRecipt             = 0x11
)

// Nrf51822 is a BLE SoC from Nordic
// https://www.nordicsemi.com/eng/Products/Bluetooth-low-energy/nRF51822
type Nrf51822 struct {
	Log                 *log.Logger
	LocalUUID           string
	Firmware            FIRMWARE
	NotificationChannel chan []byte
}

type FIRMWARE struct {
	currentBlock int
	size         int
	binary       []byte
	data         []byte
}

var (
	dfuPkt  *ble.Characteristic
	dfuCtrl *ble.Characteristic
)

func (m *Nrf51822) ExtractFirmware(filePath, bin, data string) error {
	m.Log.WithFields(log.Fields{
		"Firmware path": filePath,
		"Bin":           bin,
		"Data":          data,
	}).Debug("Extracting firmware")

	var err error

	if err = archiver.Zip.Open(path.Join(filePath, "application.zip"), filePath); err != nil {
		return err
	}

	m.Firmware.binary, err = ioutil.ReadFile(path.Join(filePath, bin))
	if err != nil {
		return err
	}

	m.Firmware.data, err = ioutil.ReadFile(path.Join(filePath, data))
	if err != nil {
		return err
	}

	m.Firmware.size = len(m.Firmware.binary)

	m.Log.WithFields(log.Fields{
		"Size": m.Firmware.size,
	}).Debug("Extracted firmware")

	return nil
}

func (m *Nrf51822) Update(client ble.Client) error {
	if err := m.subscribe(client); err != nil {
		return err
	}
	defer client.ClearSubscriptions()

	if err := m.checkFOTA(client); err != nil {
		return err
	}

	if m.Firmware.currentBlock == 0 {
		if err := m.initFOTA(client); err != nil {
			return err
		}
	}

	if err := m.transferFOTA(client); err != nil {
		return err
	}

	if err := m.validateFOTA(client); err != nil {
		return err
	}

	return m.finaliseFOTA(client)
}

func init() {
	log.SetLevel(config.GetLogLevel())

	var err error
	dfuCtrl, err = bluetooth.GetCharacteristic("000015311212efde1523785feabcd123", ble.CharWrite+ble.CharNotify, 0x0F, 0x10)
	if err != nil {
		log.Fatal(err)
	}

	descriptor, err := bluetooth.GetDescriptor("2902", 0x11)
	if err != nil {
		log.Fatal(err)
	}
	dfuCtrl.CCCD = descriptor

	dfuPkt, err = bluetooth.GetCharacteristic("000015321212efde1523785feabcd123", ble.CharWriteNR, 0x0D, 0x0E)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Initialised nRF51822 characteristics")
}

func (m *Nrf51822) subscribe(client ble.Client) error {
	if err := client.WriteDescriptor(dfuCtrl.CCCD, []byte{0x0001}); err != nil {
		return err
	}

	return client.Subscribe(dfuCtrl, false, func(b []byte) {
		m.NotificationChannel <- b
	})
}

func (m *Nrf51822) checkFOTA(client ble.Client) error {
	m.Log.Debug("Checking FOTA")

	if err := client.WriteCharacteristic(dfuCtrl, []byte{ReceivedSize}, false); err != nil {
		return err
	}

	resp, err := m.getNotification([]byte{Response, ReceivedSize, Success}, true)
	if err != nil {
		return err
	}

	m.Firmware.currentBlock, err = unpack(resp[3:])
	if err != nil {
		return err
	}

	m.Log.WithFields(log.Fields{
		"Start block": m.Firmware.currentBlock,
	}).Debug("Checked FOTA")

	return nil
}

func (m *Nrf51822) initFOTA(client ble.Client) error {
	m.Log.Debug("Initialising FOTA")

	if err := client.WriteCharacteristic(dfuCtrl, []byte{Start, 0x04}, false); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if _, err := buf.Write(make([]byte, 8)); err != nil {
		return err
	}

	if err := binary.Write(buf, binary.LittleEndian, (int32)(m.Firmware.size)); err != nil {
		return err
	}

	if err := client.WriteCharacteristic(dfuPkt, buf.Bytes(), false); err != nil {
		return err
	}

	if _, err := m.getNotification([]byte{Response, Start, Success}, true); err != nil {
		return err
	}

	if err := client.WriteCharacteristic(dfuCtrl, []byte{Initialise, 0x00}, false); err != nil {
		return err
	}

	if err := client.WriteCharacteristic(dfuPkt, m.Firmware.data, false); err != nil {
		return err
	}

	if err := client.WriteCharacteristic(dfuCtrl, []byte{Initialise, 0x01}, false); err != nil {
		return err
	}

	if _, err := m.getNotification([]byte{Response, Initialise, Success}, true); err != nil {
		return err
	}

	if err := client.WriteCharacteristic(dfuCtrl, []byte{RequestBlockRecipt, 0x64, 0x00}, false); err != nil {
		return err
	}

	if err := client.WriteCharacteristic(dfuCtrl, []byte{Receive}, false); err != nil {
		return err
	}

	m.Log.Debug("Initialised FOTA")

	return nil
}

func (m *Nrf51822) transferFOTA(client ble.Client) error {
	blockCounter := 1
	blockSize := 20

	if m.Firmware.currentBlock != 0 {
		blockCounter += (m.Firmware.currentBlock / blockSize)
	}

	m.Log.WithFields(log.Fields{
		"Progress %": m.getProgress(),
	}).Info("Transferring FOTA")

	for i := m.Firmware.currentBlock; i < m.Firmware.size; i += blockSize {
		sliceIndex := i + blockSize
		if sliceIndex > m.Firmware.size {
			sliceIndex = m.Firmware.size
		}
		block := m.Firmware.binary[i:sliceIndex]

		if err := client.WriteCharacteristic(dfuPkt, block, true); err != nil {
			return err
		}
		// time.Sleep(time.Duration(10) * time.Millisecond)

		if (blockCounter % 100) == 0 {
			resp, err := m.getNotification(nil, false)
			if err != nil {
				return err
			}

			if resp[0] != BlockRecipt {
				return fmt.Errorf("Incorrect notification received")
			}

			if m.Firmware.currentBlock, err = unpack(resp[1:]); err != nil {
				return err
			}

			if (i + blockSize) != m.Firmware.currentBlock {
				return fmt.Errorf("FOTA transer out of sync")
			}

			m.Log.WithFields(log.Fields{
				"Progress %": m.getProgress(),
			}).Info("Transferring FOTA")
		}

		blockCounter++
	}

	if _, err := m.getNotification([]byte{Response, Receive, Success}, true); err != nil {
		return err
	}

	m.Log.WithFields(log.Fields{
		"Progress %": 100,
	}).Info("Transferring FOTA")

	return nil
}

func (m *Nrf51822) validateFOTA(client ble.Client) error {
	m.Log.Debug("Validating FOTA")

	if err := m.checkFOTA(client); err != nil {
		return err
	}

	if m.Firmware.currentBlock != m.Firmware.size {
		return fmt.Errorf("Bytes received does not match binary size")
	}

	if err := client.WriteCharacteristic(dfuCtrl, []byte{Validate}, false); err != nil {
		return err
	}

	if _, err := m.getNotification([]byte{Response, Validate, Success}, true); err != nil {
		return err
	}

	m.Log.Debug("Validated FOTA")

	return nil
}

func (m Nrf51822) finaliseFOTA(client ble.Client) error {
	m.Log.Debug("Finalising FOTA")

	// Ignore the error because this command causes the micro:bit to disconnect
	client.WriteCharacteristic(dfuCtrl, []byte{Activate}, false)

	m.Log.Debug("Finalised FOTA")

	return nil
}

func (m *Nrf51822) getNotification(exp []byte, compare bool) ([]byte, error) {
	for {
		select {
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("Timed out waiting for notification")
		case resp := <-m.NotificationChannel:
			if !compare || bytes.Equal(resp[:3], exp) {
				return resp, nil
			}

			m.Log.WithFields(log.Fields{
				"[0]": fmt.Sprintf("0x%X", resp[0]),
				"[1]": fmt.Sprintf("0x%X", resp[1]),
				"[2]": fmt.Sprintf("0x%X", resp[2]),
			}).Debug("Received")

			m.Log.WithFields(log.Fields{
				"[0]": fmt.Sprintf("0x%X", exp[0]),
				"[1]": fmt.Sprintf("0x%X", exp[1]),
				"[2]": fmt.Sprintf("0x%X", exp[2]),
			}).Debug("Expected")

			return nil, fmt.Errorf("Incorrect notification received")
		}
	}
}

func (m *Nrf51822) getProgress() float32 {
	return ((float32)(m.Firmware.currentBlock) / (float32)(m.Firmware.size)) * 100.0
}

func unpack(resp []byte) (int, error) {
	var result int32
	buf := bytes.NewReader(resp)
	if err := binary.Read(buf, binary.LittleEndian, &result); err != nil {
		return 0, err
	}
	return (int)(result), nil
}
