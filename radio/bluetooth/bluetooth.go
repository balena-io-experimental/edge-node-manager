package bluetooth

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

var (
	Radio         gatt.Device
	periphChannel = make(chan gatt.Peripheral)
)

func Scan(id string, timeout time.Duration) (map[string]bool, error) {
	Radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	if err := Radio.Init(OnStateChanged); err != nil {
		return nil, err
	}

	devices := make(map[string]bool)

	for {
		select {
		case <-time.After(timeout * time.Second):
			Radio.StopScanning()
			return devices, nil
		case onlineDevice := <-periphChannel:
			if onlineDevice.Name() == id {
				devices[onlineDevice.ID()] = true
			}
		}
	}
}

func Online(id string, timeout time.Duration) (bool, error) {
	Radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	if err := Radio.Init(OnStateChanged); err != nil {
		return false, err
	}

	for {
		select {
		case <-time.After(timeout * time.Second):
			Radio.StopScanning()
			return false, nil
		case onlineDevice := <-periphChannel:
			if onlineDevice.ID() == id {
				Radio.StopScanning()
				return true, nil
			}
		}
	}
}

func Print(periph gatt.Peripheral) error {
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

func GetChar(serUUID, charUUID, descUUID string, props gatt.Property, h, vh uint16) (*gatt.Characteristic, error) {
	serviceUUID, err := gatt.ParseUUID(serUUID)
	if err != nil {
		return &gatt.Characteristic{}, err
	}

	characteristicUUID, err := gatt.ParseUUID(charUUID)
	if err != nil {
		return &gatt.Characteristic{}, err
	}

	service := gatt.NewService(serviceUUID)
	characteristic := gatt.NewCharacteristic(characteristicUUID, service, props, h, vh)

	if descUUID == "" {
		return characteristic, nil
	}

	descriptorUUID, err := gatt.ParseUUID(descUUID)
	if err != nil {
		return &gatt.Characteristic{}, err
	}

	descriptor := gatt.NewDescriptor(descriptorUUID, 17, characteristic)
	characteristic.SetDescriptor(descriptor)

	return characteristic, nil
}

func GetName(periph gatt.Peripheral) (string, error) {
	characteristic, err := GetChar("1800", "2a00", "", gatt.CharRead+gatt.CharWrite, 2, 3)
	if err != nil {
		return "", err
	}

	byte, err := periph.ReadCharacteristic(characteristic)
	if err != nil {
		return "", err
	}

	return string(byte), nil
}

func OnStateChanged(device gatt.Device, state gatt.State) {
	log.WithFields(log.Fields{
		"State": state,
	}).Debug("State changed")
	switch state {
	case gatt.StatePoweredOn:
		device.Scan([]gatt.UUID{}, false)
	default:
		device.StopScanning()
	}
}

func onPeriphDiscovered(periph gatt.Peripheral, adv *gatt.Advertisement, rssi int) {
	periphChannel <- periph
}

func init() {
	log.SetLevel(config.GetLogLevel())

	var err error
	if Radio, err = gatt.NewDevice(option.DefaultClientOptions...); err != nil {
		log.WithFields(log.Fields{
			"Options": option.DefaultClientOptions,
			"Error":   err,
		}).Fatal("Unable to create a new gatt device")
	}

	log.WithFields(log.Fields{
		"Options": option.DefaultClientOptions,
	}).Debug("Created new gatt device")
}
