package device

import "log"

type operation byte

const (
	START_DFU                   operation = 0x01 //Can I use iota here?
	INITIALISE_DFU              operation = 0x02
	RECEIVE_FIRMWARE_IMAGE      operation = 0x03
	VALIDATE_FIRMWARE_IMAGE     operation = 0x04
	ACTIVATE_FIRMWARE_AND_RESET operation = 0x05
	SYSTEM_RESET                operation = 0x06
	REPORT_RECEIVED             operation = 0x07
	PKT_RCPT_NOTIF_REQ          operation = 0x08
	RESPONSE                    operation = 0x10
	PKT_RCPT_NOTIF              operation = 0x11
)

type procedure byte

const (
	START          procedure = 0x01 //Can I use iota here?
	INITIALISE     procedure = 0x02
	RECEIVE_APP    procedure = 0x03
	VALIDATE       procedure = 0x04
	IMAGE_SIZE_REQ procedure = 0x07
	PKT_RCPT_REQ   procedure = 0x08
)

type response byte

const (
	SUCCESS          response = 0x01 //Can I use iota here?
	INVALID_STATE    response = 0x02
	NOT_SUPPORTED    response = 0x03
	DATA_SIZE        response = 0x04
	CRC_ERROR        response = 0x05
	OPERATION_FAILED response = 0x06
)

type updateType byte

const (
	SOFT_DEVICE            updateType = 0x01 //Can I use iota here?
	BOOTLOADER             updateType = 0x02
	SOFT_DEVICE_BOOTLOADER updateType = 0x03
	APPLICATION            updateType = 0x04
)

type handle byte

const (
	CONTROL handle = 0x10
	PACKET  handle = 0x0E
)

type Nrf51822 struct {
	*Device
	packetSize     int
	nameHandle     byte
	identifyHandle byte
	restartHandle  byte
}

func (d Nrf51822) GetDevice() *Device {
	return d.Device
}

func (d Nrf51822) Update(application, commit string) error {
	log.Println("Updating NRF51822")
	return nil
}

func (d Nrf51822) Scan() ([]string, error) {
	return d.GetDevice().Radio.Scan(d.ApplicationUUID, 10)
}

func (d Nrf51822) Online() (bool, error) {
	return d.GetDevice().Radio.Online(d.LocalUUID, 10)
}

func (d Nrf51822) Identify() error {
	return nil
	//return d.GetDevice().Radio.WriteCharacteristic(d.identifyHandle, 0x01) //How to get access to underlying type
}

func (d Nrf51822) Restart() error {
	return nil
	//return d.GetDevice().Radio.WriteCharacteristic(d.restartHandle, 0x01) //How to get access to underlying type
}
