package serial

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/kylelemons/gousb/usb"
	"github.com/resin-io/edge-node-manager/config"
)

type Device struct {
	Descriptor   *usb.Descriptor
	Manufacturer string
	Product      string
	Serial       string
}

func Scan(vendorID, productID usb.ID) (map[string]bool, error) {
	serials := make(map[string]bool)

	devices, err := scan(vendorID, productID)
	if err != nil {
		return serials, err
	}

	for key, _ := range devices {
		serials[key] = true
	}

	return serials, err
}

func Online(serial string, vendorID, productID usb.ID) (bool, error) {
	devices, err := scan(vendorID, productID)
	if err != nil {
		return false, err
	}

	_, online := devices[serial]
	return online, nil
}

func GetBusAddress(serial string, vendorID, productID usb.ID) (uint8, uint8, error) {
	devices, err := scan(vendorID, productID)
	if err != nil {
		return 0, 0, err
	}

	device, ok := devices[serial]
	if !ok {
		return 0, 0, fmt.Errorf("Device not online")
	}

	return device.Descriptor.Bus, device.Descriptor.Address, nil
}

func init() {
	log.SetLevel(config.GetLogLevel())

	log.Debug("Created a new serial device")
}

func scan(vendorID, productID usb.ID) (map[string]Device, error) {
	devices := make(map[string]Device)

	ctx := usb.NewContext()
	defer ctx.Close()

	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		return (desc.Vendor == vendorID && desc.Product == productID)
	})

	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	if err != nil {
		return devices, err
	}

	for _, dev := range devs {
		manufacturer, err := dev.GetStringDescriptor(1)
		if err != nil {
			return devices, err
		}
		product, err := dev.GetStringDescriptor(2)
		if err != nil {
			return devices, err
		}
		serial, err := dev.GetStringDescriptor(3)
		if err != nil {
			return devices, err
		}

		devices[serial] = Device{
			Descriptor:   dev.Descriptor,
			Manufacturer: manufacturer,
			Product:      product,
			Serial:       serial,
		}
	}

	return devices, err
}
