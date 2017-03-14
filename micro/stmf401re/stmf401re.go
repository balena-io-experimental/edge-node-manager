package stmf401re

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/kylelemons/gousb/usb"
	"github.com/resin-io/edge-node-manager/config"
)

type Stmf401re struct {
	Log       *log.Logger
	LocalUUID string
	VendorID  usb.ID
	ProductID usb.ID
}

func (m *Stmf401re) Update(bus, address uint8, filePath string) error {
	busAddressKey := "STLINK_DEVICE"
	busAddressValue := fmt.Sprintf("%03d:%03d", bus, address)
	command := "st-flash"
	args := []string{"--format", "ihex", "write", path.Join(filePath, "main.hex")}

	m.Log.WithFields(log.Fields{
		"Key":     busAddressKey,
		"Value":   busAddressValue,
		"Command": command,
		"Args":    args,
	}).Debug("Update command")

	if err := os.Setenv(busAddressKey, busAddressValue); err != nil {
		return err
	}

	if err := exec.Command(command, args...).Run(); err != nil {
		return err
	}

	return nil
}

func init() {
	log.SetLevel(config.GetLogLevel())
}
