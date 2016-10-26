package nrf51822

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/mholt/archiver"
	"github.com/paypal/gatt"
	"github.com/resin-io/edge-node-manager/radio/bluetooth"

	log "github.com/Sirupsen/logrus"
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
	LocalUUID        string
	Fota             FOTA
	ConnectedChannel chan bool
	RestartChannel   chan bool
	ErrChannel       chan error
}

type FOTA struct {
	progress     float32
	currentBlock int
	binary       []byte
	data         []byte
	size         int
}

func (m *Nrf51822) ExtractFirmware(filePath, bin, dat string) error {
	if err := archiver.Unzip(path.Join(filePath, "application.zip"), filePath); err != nil {
		return err
	}

	var err error

	m.Fota.binary, err = ioutil.ReadFile(path.Join(filePath, bin))
	if err != nil {
		return err
	}

	m.Fota.data, err = ioutil.ReadFile(path.Join(filePath, dat))
	if err != nil {
		return err
	}

	m.Fota.size = len(m.Fota.binary)

	return nil
}

func (m *Nrf51822) ProcessRequest(f func(gatt.Peripheral, error)) error {
	bluetooth.Radio.Handle(
		gatt.PeripheralDiscovered(m.OnPeriphDiscovered),
		gatt.PeripheralConnected(f),
		gatt.PeripheralDisconnected(m.OnPeriphDisconnected),
	)
	if err := bluetooth.Radio.Init(bluetooth.OnStateChanged); err != nil {
		return err
	}

	var savedErr error
	for {
		select {
		case savedErr = <-m.ErrChannel:
		case connected := <-m.ConnectedChannel:
			if connected {
				log.Debug("Connected")
			} else {
				log.Debug("Disconnected")
				return savedErr
			}
		}
	}
}

func (m *Nrf51822) OnPeriphDiscovered(periph gatt.Peripheral, adv *gatt.Advertisement, rssi int) {
	if periph.ID() != m.LocalUUID {
		return
	}

	periph.Device().StopScanning()
	periph.Device().Connect(periph)
}

func (m *Nrf51822) OnPeriphDisconnected(periph gatt.Peripheral, err error) {
	m.ConnectedChannel <- false
}

func (m *Nrf51822) UpdateOnPeriphConnected(periph gatt.Peripheral, err error) {
	defer periph.Device().CancelConnection(periph)

	m.ConnectedChannel <- true

	if err := periph.SetMTU(500); err != nil {
		m.ErrChannel <- err
		return
	}

	if err := m.checkFOTA(periph); err != nil {
		m.ErrChannel <- err
		return
	}

	if m.Fota.currentBlock == 0 {
		if err := m.initFOTA(periph); err != nil {
			m.ErrChannel <- err
			return
		}
	}

	if err := m.transferFOTA(periph); err != nil {
		m.ErrChannel <- err
		return
	}

	if err := m.validateFOTA(periph); err != nil {
		m.ErrChannel <- err
		return
	}

	if err := m.finaliseFOTA(periph); err != nil {
		m.ErrChannel <- err
		return
	}
}

func (m *Nrf51822) EnableCCCD(periph gatt.Peripheral) error {
	characteristic, err := bluetooth.GetChar("000015301212efde1523785feabcd123", "000015311212efde1523785feabcd123", "2902", gatt.CharWrite+gatt.CharNotify, 15, 16)
	if err != nil {
		return err
	}

	return periph.WriteDescriptor(characteristic.Descriptor(), []byte{Start, 0x00})
}

func (m *Nrf51822) WriteDFUControlPoint(periph gatt.Peripheral, value []byte, noRsp bool) error {
	characteristic, err := bluetooth.GetChar("000015301212efde1523785feabcd123", "000015311212efde1523785feabcd123", "2902", gatt.CharWrite+gatt.CharNotify, 15, 16)
	if err != nil {
		return err
	}

	return periph.WriteCharacteristic(characteristic, value, noRsp)
}

func (m *Nrf51822) checkFOTA(periph gatt.Peripheral) error {
	log.Debug("Checking FOTA")

	if err := m.EnableCCCD(periph); err != nil {
		return err
	}

	resp, err := m.notifyDFUControlPoint(periph, []byte{ReceivedSize})
	if err != nil {
		return err
	}

	if resp[0] != Response || resp[1] != ReceivedSize || resp[2] != Success {
		return fmt.Errorf("Incorrect notification received")
	}

	m.Fota.currentBlock, err = m.unpack(resp[3:])
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Start block": m.Fota.currentBlock,
	}).Debug("Checked FOTA")

	return err
}

func (m *Nrf51822) initFOTA(periph gatt.Peripheral) error {
	log.Debug("Initialising FOTA")

	if err := m.EnableCCCD(periph); err != nil {
		return err
	}

	if err := m.WriteDFUControlPoint(periph, []byte{Start, 0x04}, false); err != nil {
		return err
	}

	size, err := m.pack()
	if err != nil {
		return err
	}

	resp, err := m.notifyDFUPacket(periph, size)
	if err != nil {
		return err
	}

	if resp[0] != Response || resp[1] != Start || resp[2] != Success {
		return fmt.Errorf("Incorrect notification received")
	}

	if err = m.WriteDFUControlPoint(periph, []byte{Initialise, 0x00}, false); err != nil {
		return err
	}

	if err = m.writeDFUPacket(periph, m.Fota.data, false); err != nil {
		return err
	}

	resp, err = m.notifyDFUControlPoint(periph, []byte{Initialise, 0x01})
	if err != nil {
		return err
	}

	if resp[0] != Response || resp[1] != Initialise || resp[2] != Success {
		return fmt.Errorf("Incorrect notification received")
	}

	if err = m.WriteDFUControlPoint(periph, []byte{RequestBlockRecipt, 0x64, 0x00}, false); err != nil {
		return err
	}

	if err = m.WriteDFUControlPoint(periph, []byte{Receive}, false); err != nil {
		return err
	}

	log.Debug("Initialised FOTA")

	return nil
}

func (m *Nrf51822) transferFOTA(periph gatt.Peripheral) error {
	blockCounter := 1
	blockSize := 20
	if m.Fota.currentBlock != 0 {
		// Set block counter to the current block, this is used to resume FOTA if
		// the previous FOTA was cancelled/failed mid way though
		blockCounter += (m.Fota.currentBlock / blockSize)
	}

	m.Fota.progress = ((float32)(m.Fota.currentBlock) / (float32)(m.Fota.size)) * 100.0

	log.WithFields(log.Fields{
		"Block counter": blockCounter,
		"Progress %":    m.Fota.progress,
	}).Debug("Transferring FOTA")

	notifyChannel, err := m.initNotify(periph)
	if err != nil {
		return err
	}

	for i := m.Fota.currentBlock; i < m.Fota.size; i += blockSize {
		sliceIndex := i + blockSize
		if sliceIndex > m.Fota.size {
			// Limit the slice to Fota.Size to avoid extra zeros being tagged on the end of the block
			sliceIndex = m.Fota.size
		}
		block := m.Fota.binary[i:sliceIndex]

		if err = m.writeDFUPacket(periph, block, false); err != nil {
			return err
		}

		if (blockCounter % 100) == 0 {
			resp, err := m.timeoutNotify(notifyChannel)
			if err != nil {
				return err
			}

			if resp[0] != BlockRecipt {
				return fmt.Errorf("Incorrect notification received")
			}

			currentBlock, err := m.unpack(resp[1:])
			if err != nil {
				return err
			}

			if (i + blockSize) != currentBlock {
				return fmt.Errorf("FOTA transer out of sync")
			}

			m.Fota.progress = ((float32)(currentBlock) / (float32)(m.Fota.size)) * 100.0

			log.WithFields(log.Fields{
				"Block counter": blockCounter,
				"Progress %":    m.Fota.progress,
			}).Debug("Transferring FOTA")
		}

		blockCounter++
	}

	m.Fota.progress = 100

	resp, err := m.timeoutNotify(notifyChannel)
	if err != nil {
		return err
	}

	if resp[0] != Response || resp[1] != Receive || resp[2] != Success {
		return fmt.Errorf("Incorrect notification received")
	}

	log.WithFields(log.Fields{
		"Block counter": blockCounter,
		"Progress %":    m.Fota.progress,
	}).Debug("Transferred FOTA")

	return nil
}

func (m *Nrf51822) validateFOTA(periph gatt.Peripheral) error {
	log.Debug("Validating FOTA")

	if err := m.checkFOTA(periph); err != nil {
		return err
	}

	if m.Fota.currentBlock != m.Fota.size {
		return fmt.Errorf("Bytes received does not match binary size")
	}

	resp, err := m.notifyDFUControlPoint(periph, []byte{Validate})
	if err != nil {
		return err
	}

	if resp[0] != Response || resp[1] != Validate || resp[2] != Success {
		return fmt.Errorf("Incorrect notification received")
	}

	log.Debug("Validated FOTA")

	return nil
}

func (m Nrf51822) finaliseFOTA(periph gatt.Peripheral) error {
	log.Debug("Finalising FOTA")

	if err := m.WriteDFUControlPoint(periph, []byte{Activate}, false); err != nil {
		return err
	}

	log.Debug("Finalised FOTA")

	return nil
}

func (m *Nrf51822) writeDFUPacket(periph gatt.Peripheral, value []byte, noRsp bool) error {
	characteristic, err := bluetooth.GetChar("000015301212efde1523785feabcd123", "000015321212efde1523785feabcd123", "2902", gatt.CharWriteNR, 13, 14)
	if err != nil {
		return err
	}

	return periph.WriteCharacteristic(characteristic, value, noRsp)
}

func (m *Nrf51822) notifyDFUControlPoint(periph gatt.Peripheral, value []byte) ([]byte, error) {
	notifyChannel, err := m.initNotify(periph)
	if err != nil {
		return nil, err
	}

	if err = m.WriteDFUControlPoint(periph, value, false); err != nil {
		return nil, err
	}

	return m.timeoutNotify(notifyChannel)

}

func (m *Nrf51822) notifyDFUPacket(periph gatt.Peripheral, value []byte) ([]byte, error) {
	notifyChannel, err := m.initNotify(periph)
	if err != nil {
		return nil, err
	}

	if err = m.writeDFUPacket(periph, value, false); err != nil {
		return nil, err
	}

	return m.timeoutNotify(notifyChannel)
}

func (m *Nrf51822) initNotify(periph gatt.Peripheral) (chan []byte, error) {
	notifyChannel := make(chan []byte)

	characteristic, err := bluetooth.GetChar("000015301212efde1523785feabcd123", "000015311212efde1523785feabcd123", "2902", gatt.CharWrite+gatt.CharNotify, 15, 16)
	if err != nil {
		return notifyChannel, err
	}

	callback := func(c *gatt.Characteristic, b []byte, err error) {
		notifyChannel <- b
	}

	err = periph.SetNotifyValue(characteristic, callback)

	return notifyChannel, err
}

func (m *Nrf51822) timeoutNotify(notifyChannel chan []byte) ([]byte, error) {
	for {
		select {
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("Timed out waiting for notification")
		case resp := <-notifyChannel:
			log.WithFields(log.Fields{
				"[0]": resp[0],
				"[1]": resp[1],
				"[2]": resp[2],
			}).Debug("Notification")
			return resp, nil
		}
	}
}

func (m *Nrf51822) unpack(resp []byte) (int, error) {
	var result int32
	buf := bytes.NewReader(resp)
	if err := binary.Read(buf, binary.LittleEndian, &result); err != nil {
		return 0, err
	}
	return (int)(result), nil
}

func (m *Nrf51822) pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Pad the buffer with 8 zeroed bytes
	if _, err := buf.Write(make([]byte, 8)); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, (int32)(m.Fota.size)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
