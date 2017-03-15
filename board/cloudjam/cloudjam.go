package cloudjam

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/kylelemons/gousb/usb"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/micro/stmf401re"
	"github.com/resin-io/edge-node-manager/radio/serial"
)

type Cloudjam struct {
	Log   *log.Logger
	Micro stmf401re.Stmf401re
}

var (
	vendorID  usb.ID
	productID usb.ID
)

func (b Cloudjam) Update(path string) error {
	b.Log.Info("Starting update")

	bus, address, err := serial.GetBusAddress(b.Micro.LocalUUID, b.Micro.VendorID, b.Micro.ProductID)
	if err != nil {
		return err
	}

	if err := b.Micro.Update(bus, address, path); err != nil {
		return err
	}

	b.Log.Info("Finished update")

	return nil
}

func (b Cloudjam) Scan(applicationUUID int) (map[string]bool, error) {
	return serial.Scan(vendorID, productID)
}

func (b Cloudjam) Online() (bool, error) {
	return serial.Online(b.Micro.LocalUUID, b.Micro.VendorID, b.Micro.ProductID)
}

func (b Cloudjam) Restart() error {
	b.Log.Info("Restarting...")
	return fmt.Errorf("Restart not implemented")
}

func (b Cloudjam) Identify() error {
	b.Log.Info("Identifying...")
	return fmt.Errorf("Identify not implemented")
}

func (b Cloudjam) UpdateConfig(config interface{}) error {
	b.Log.WithFields(log.Fields{
		"Config": config,
	}).Info("Updating config...")
	return fmt.Errorf("Update config not implemented")
}

func (b Cloudjam) UpdateEnvironment(config interface{}) error {
	b.Log.WithFields(log.Fields{
		"Config": config,
	}).Info("Updating environment...")
	return fmt.Errorf("Update environment not implemented")
}

func init() {
	log.SetLevel(config.GetLogLevel())

	vendorID = 0x0483
	productID = 0x374b

	log.Debug("Initialised Cloud-Jam ID's")
}
