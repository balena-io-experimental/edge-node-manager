// +build test ignore

package database

import (
	log "github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/config"
)

func init() {
	log.SetLevel(config.GetLogLevel())

	directory := config.GetDbDir()
	name := config.GetDbName()

	initialise(directory, name)

	log.WithFields(log.Fields{
		"Path": dbPath,
	}).Debug("Created database path")
}
