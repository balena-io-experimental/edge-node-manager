package device

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/paypal/gatt"

	"github.com/josephroberts/edge-node-manager/radio/bluetooth"
)

// ALL THESE TYPES WILL BE USED LATER ON @LORENZO

type operation byte

/*
 * Operations defined by device
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
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
 * Responses defined by device
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
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
 * Update type defined by device
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
 */
const (
	SOFT_DEVICE            updateType = 0x01
	BOOTLOADER                        = 0x02
	SOFT_DEVICE_BOOTLOADER            = 0x03
	APPLICATION                       = 0x04
)

type handle byte

/*
 * Handles defined by device
 * https://infocenter.nordicsemi.com/index.jsp?topic=%2Fcom.nordic.infocenter.sdk5.v11.0.0%2Fbledfu_transport_bleservice.html&anchor=ota_spec_control_state
 */
const (
	CONTROL handle = 0x10
	PACKET         = 0x0E
)

// Defined by firmware running on the device
const (
	PACKET_SIZE     int  = 20
	NAME_HANDLE     byte = 0x03
	IDENTIFY_HANDLE byte = 0x16
	RESTART_HANDLE  byte = 0x18
)

type Nrf51822 Device

func (d Nrf51822) String() string {
	return (Device)(d).String()
}

func (d Nrf51822) Update(application, commit string) error {
	log.WithFields(log.Fields{
		"Device":           d,
		"Application UUID": application,
		"Commit":           commit,
	}).Debug("Update")
	return nil
}

func (d Nrf51822) Online() (bool, error) {
	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Online")

	bluetooth.Radio.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)
	if err := bluetooth.Radio.Init(onStateChanged); err != nil {
		return false, err
	}

	<-done
	fmt.Println("Done")

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

func (d Nrf51822) startBootloader() error {
	return nil
}

func (d Nrf51822) checkFOTA() error {
	return nil
}

func (d Nrf51822) initFOTA() error {
	return nil
}

func (d Nrf51822) transferFOTA() error {
	return nil
}

func (d Nrf51822) validateFOTA() error {
	return nil
}

///

var done = make(chan struct{})

func onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	fmt.Println(p.ID())

	id := "EE:50:F0:F8:3C:FF"
	if strings.ToUpper(p.ID()) != id {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	p.Device().StopScanning()

	fmt.Printf("\nPeripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
	fmt.Println("  Local Name        =", a.LocalName)
	fmt.Println("  TX Power Level    =", a.TxPowerLevel)
	fmt.Println("  Manufacturer Data =", a.ManufacturerData)
	fmt.Println("  Service Data      =", a.ServiceData)
	fmt.Println("")

	p.Device().Connect(p)
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Println("Connected")
	defer p.Device().CancelConnection(p)

	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}

	// Discovery services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
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

			// Read the characteristic, if possible.
			if (c.Properties() & gatt.CharRead) != 0 {
				b, err := p.ReadCharacteristic(c)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Discovery descriptors
			ds, err := p.DiscoverDescriptors(nil, c)
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

				// Read descriptor (could fail, if it's not readable)
				b, err := p.ReadDescriptor(d)
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
				if err := p.SetNotifyValue(c, f); err != nil {
					fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
					continue
				}
			}

		}
		fmt.Println()
	}

	fmt.Printf("Waiting for 5 seconds to get some notifiations, if any.\n")
	time.Sleep(5 * time.Second)
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
	close(done)
}
