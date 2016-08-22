package device

import "log"

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

/* Responses defined by device
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

/* Update type defined by device
https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
*/
const (
	SOFT_DEVICE            updateType = 0x01
	BOOTLOADER                        = 0x02
	SOFT_DEVICE_BOOTLOADER            = 0x03
	APPLICATION                       = 0x04
)

type handle byte

/* Handles defined by device
https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state*/
const (
	CONTROL handle = 0x10
	PACKET         = 0x0E
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
