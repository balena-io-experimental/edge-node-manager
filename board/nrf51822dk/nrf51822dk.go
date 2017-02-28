package nrf51822dk

import (
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/currantlabs/ble"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/micro/nrf51822"
	"github.com/resin-io/edge-node-manager/radio/bluetooth"
)

type Nrf51822dk struct {
	Log   *log.Logger
	Micro nrf51822.Nrf51822
}

var dfu *ble.Characteristic

func (b Nrf51822dk) Update(path string) error {
	b.Log.Info("Starting update")

	if err := b.Micro.ExtractFirmware(path, "nrf51422_xxac_s130.bin", "nrf51422_xxac_s130.dat"); err != nil {
		return err
	}

	name, err := bluetooth.GetName(b.Micro.LocalUUID, 10)
	if err != nil {
		return err
	}

	if name != "DfuTarg" {
		b.Log.Debug("Starting bootloader")

		client, err := bluetooth.Connect(b.Micro.LocalUUID, 10)
		if err != nil {
			return err
		}

		if err = client.WriteDescriptor(dfu.CCCD, []byte{0x001}); err != nil {
			return err
		}

		if err = client.WriteCharacteristic(dfu, []byte{nrf51822.Start, 0x04}, false); err != nil {
			return err
		}

		if err = bluetooth.Disconnect(client); err != nil {
			return err
		}

		b.Log.Debug("Started bootloader")
	} else {
		b.Log.Debug("Bootloader already started")
	}

	client, err := bluetooth.Connect(b.Micro.LocalUUID, 10)
	if err != nil {
		return err
	}

	if err := b.Micro.Update(client); err != nil {
		return err
	}

	if err := bluetooth.Disconnect(client); err != nil {
		return err
	}

	b.Log.Info("Finished update")

	return nil
}

func (b Nrf51822dk) Scan(applicationUUID int) (map[string]bool, error) {
	return bluetooth.Scan(strconv.Itoa(applicationUUID), 10)
}

func (b Nrf51822dk) Online() (bool, error) {
	return bluetooth.Online(b.Micro.LocalUUID, 10)
}

func (b Nrf51822dk) Restart() error {
	b.Log.Info("Restarting...")
	return fmt.Errorf("Restart not implemented")
}

func (b Nrf51822dk) Identify() error {
	b.Log.Info("Identifying...")
	return fmt.Errorf("Identify not implemented")
}

func (b Nrf51822dk) UpdateConfig(config interface{}) error {
	b.Log.WithFields(log.Fields{
		"Config": config,
	}).Info("Updating config...")
	return fmt.Errorf("Update config not implemented")
}

func (b Nrf51822dk) UpdateEnvironment(config interface{}) error {
	b.Log.WithFields(log.Fields{
		"Config": config,
	}).Info("Updating environment...")
	return fmt.Errorf("Update environment not implemented")
}

func init() {
	log.SetLevel(config.GetLogLevel())

	var err error
	dfu, err = bluetooth.GetCharacteristic("000015311212efde1523785feabcd123", ble.CharWrite+ble.CharNotify, 0x0F, 0x10)
	if err != nil {
		log.Fatal(err)
	}

	descriptor, err := bluetooth.GetDescriptor("2902", 0x11)
	if err != nil {
		log.Fatal(err)
	}
	dfu.CCCD = descriptor

	log.Debug("Initialised nRF51822-DK characteristics")
}
