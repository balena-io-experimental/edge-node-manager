package process

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/application"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/process/status"
)

var (
	delay         time.Duration
	CurrentStatus status.Status
	TargetStatus  status.Status
)

// Run processes the application, checking for new commits, provisioning and updating devices
func Run(a *application.Application) []error {
	log.Info("----------------------------------------------------------------------------------------------------")

	// Pause the process if necessary
	pause()

	// Validate application to ensure the board type has been set
	if a.BoardType == "" {
		log.WithFields(log.Fields{
			"Application": a.Name,
			"Error":       "Application board type not set",
		}).Warn("Processing application")
		return nil
	}

	// Print application info
	log.WithFields(log.Fields{
		"Application":       a.Name,
		"Number of devices": len(a.Devices),
	}).Info("Processing application")

	// Get all online devices associated with this application
	if err := a.GetOnlineDevices(); err != nil {
		return []error{err}
	}

	// Provision non-provisoned online devices associated with this application
	if errs := a.ProvisionDevices(); errs != nil {
		return errs
	}

	// Set the status of all offline provisioned devices associated with this application to OFFLINE
	if errs := a.SetOfflineDeviceStatus(); errs != nil {
		return errs
	}

	// Update firmware for all online devices associated with this application
	if errs := a.UpdateOnlineDevices(); errs != nil {
		return errs
	}

	// Update config for all online devices associated with this application
	if errs := a.UpdateConfigOnlineDevices(); errs != nil {
		return errs
	}

	// Update environment for all online devices associated with this application
	if errs := a.UpdateEnvironmentOnlineDevices(); errs != nil {
		return errs
	}

	// Handle device flags
	if err := a.HandleFlags(); err != nil {
		return []error{err}
	}

	// Put all provisioned devices associated with this application
	if err := a.PutDevices(); err != nil {
		return []error{err}
	}

	return nil
}

func init() {
	log.SetLevel(config.GetLogLevel())

	var err error
	if delay, err = config.GetPauseDelay(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load pause delay")
	}

	CurrentStatus = status.RUNNING
	TargetStatus = status.RUNNING

	log.WithFields(log.Fields{
		"Pause delay": delay,
	}).Debug("Initialise process")
}

func pause() {
	if TargetStatus != status.PAUSED {
		return
	}

	CurrentStatus = status.PAUSED
	log.WithFields(log.Fields{
		"Status": CurrentStatus,
	}).Info("Process status")

	for TargetStatus == status.PAUSED {
		time.Sleep(delay * time.Second)
	}

	CurrentStatus = status.RUNNING
	log.WithFields(log.Fields{
		"Status": CurrentStatus,
	}).Info("Process status")
}
