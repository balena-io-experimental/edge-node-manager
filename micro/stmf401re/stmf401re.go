package stmf401re

import (
	log "github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/config"
)

type Stmf401re struct {
	Log       *log.Logger
	LocalUUID string
	Firmware  FIRMWARE
}

type FIRMWARE struct {
	currentBlock int
	size         int
	binary       []byte
	data         []byte
}

func (m *Stmf401re) Update() error {
	return nil
}

func init() {
	log.SetLevel(config.GetLogLevel())
}
