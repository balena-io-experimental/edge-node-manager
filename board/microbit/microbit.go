package microbit

import (
	"fmt"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/currantlabs/ble"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/micro/nrf51822"
	"github.com/resin-io/edge-node-manager/radio/bluetooth"
)

type Microbit struct {
	Log   *log.Logger
	Micro nrf51822.Nrf51822
}

var (
	dfu          *ble.Characteristic
	shortTimeout time.Duration
)

func (b Microbit) Update(path string) error {
	b.Log.Info("Starting update")

	if err := b.Micro.ExtractFirmware(path, "micro-bit.bin", "micro-bit.dat"); err != nil {
		return err
	}

	name, err := bluetooth.GetName(b.Micro.LocalUUID)
	if err != nil {
		return err
	}

	if name != "DfuTarg" {
		b.Log.Debug("Starting bootloader")

		client, err := bluetooth.Connect(b.Micro.LocalUUID)
		if err != nil {
			return err
		}

		// Ignore the error because this command causes the device to disconnect
		bluetooth.WriteCharacteristic(client, dfu, []byte{nrf51822.Start}, false)

		// Give the device time to disconnect
		time.Sleep(shortTimeout)

		b.Log.Debug("Started bootloader")
	} else {
		b.Log.Debug("Bootloader already started")
	}

	client, err := bluetooth.Connect(b.Micro.LocalUUID)
	if err != nil {
		return err
	}

	if err := b.Micro.Update(client); err != nil {
		return err
	}

	b.Log.Info("Finished update")

	return nil
}

func (b Microbit) Scan(applicationUUID int) (map[string]struct{}, error) {
	id := "BBC micro:bit [" + strconv.Itoa(applicationUUID) + "]"
	return bluetooth.Scan(id)
}

func (b Microbit) Online() (bool, error) {
	return bluetooth.Online(b.Micro.LocalUUID)
}

func (b Microbit) Restart() error {
	b.Log.Info("Restarting...")
	return fmt.Errorf("Restart not implemented")
}

func (b Microbit) Identify() error {
	b.Log.Info("Identifying...")
	return fmt.Errorf("Identify not implemented")
}

func (b Microbit) UpdateConfig(config interface{}) error {
	b.Log.WithFields(log.Fields{
		"Config": config,
	}).Info("Updating config...")
	return fmt.Errorf("Update config not implemented")
}

func (b Microbit) UpdateEnvironment(config interface{}) error {
	b.Log.WithFields(log.Fields{
		"Config": config,
	}).Info("Updating environment...")
	return fmt.Errorf("Update environment not implemented")
}

func init() {
	log.SetLevel(config.GetLogLevel())

	var err error
	if shortTimeout, err = config.GetShortBluetoothTimeout(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load bluetooth timeout")
	}

	dfu, err = bluetooth.GetCharacteristic("e95d93b1251d470aa062fa1922dfa9a8", ble.CharRead+ble.CharWrite, 0x0D, 0x0E)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Initialised micro:bit characteristics")
}
