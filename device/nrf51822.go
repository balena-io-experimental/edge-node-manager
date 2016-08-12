package device

import (
	"log"

	"github.com/josephroberts/edge-node-manager/radio/bluetooth"
)

type operation byte

/*
Operations defined by device
https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
*/
const (
	START_DFU                   operation = 0x01
	INITIALISE_DFU                        = 0x02
	RECEIVE_FIRMWARE_IMAGE                = 0x03
	VALIDATE_FIRMWARE_IMAGE               = 0x04
	ACTIVATE_FIRMWARE_AND_RESET           = 0x05
	SYSTEM_RESET                          = 0x06
	REPORT_RECEIVED_IMG_SIZE              = 0x07
	PKT_RCPT_NOTIF_REQ                    = 0x08
	RESPONSE                              = 0x10
	PKT_RCPT_NOTIF                        = 0x11
)

type response byte

/*
Responses defined by device
https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
*/
const (
	SUCCESS          response = 0x01
	INVALID_STATE             = 0x02
	NOT_SUPPORTED             = 0x03
	DATA_SIZE                 = 0x04
	CRC_ERROR                 = 0x05
	OPERATION_FAILED          = 0x06
)

type updateType byte

/*
Update type defined by device
https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
*/
const (
	SOFT_DEVICE            updateType = 0x01
	BOOTLOADER                        = 0x02
	SOFT_DEVICE_BOOTLOADER            = 0x03
	APPLICATION                       = 0x04
)

type handle byte

/*
Handles defined by device
https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
*/
const (
	CONTROL handle = 0x10
	PACKET         = 0x0E
)

/*
Defined by firmware running on the device
*/
const (
	PACKET_SIZE     int  = 20
	NAME_HANDLE     byte = 0x03
	IDENTIFY_HANDLE byte = 0x16
	RESTART_HANDLE  byte = 0x18
)

type Nrf51822 struct {
	*Device
}

func (d Nrf51822) GetDevice() *Device {
	return d.Device
}

func (d Nrf51822) Update(application, commit string) error {
	log.Println("Updating NRF51822")
	return nil
}

func (d Nrf51822) Online() (bool, error) {
	log.Println("Checking if NRF51822 online")
	return bluetooth.Online(d.LocalUUID, 10)
}

func (d Nrf51822) Identify() error {
	log.Println("Identifying NRF51822")
	return bluetooth.WriteCharacteristic(IDENTIFY_HANDLE, 0x01, false)
}

func (d Nrf51822) Restart() error {
	log.Println("Restarting NRF51822")
	return bluetooth.WriteCharacteristic(RESTART_HANDLE, 0x01, false)
}
