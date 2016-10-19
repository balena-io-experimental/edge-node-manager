package process

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
)

// Status defines the process statuses
type Status string

const (
	RUNNING Status = "RUNNING"
	PAUSED         = "PAUSED"
)

type status struct {
	CurrentStatus Status `json:"current"`
	TargetStatus  Status `json:"target"`
}

var (
	delay time.Duration
	State = status{
		CurrentStatus: RUNNING,
		TargetStatus:  RUNNING,
	}
)

// Run processes the application, checking for new commits, provisioning and updating devices
func Run(a *application.Application) []error {
	log.Info("----------------------------------------------------------------------------------------------------")

	// Pause the process if necessary
	if State.TargetStatus == PAUSED {
		pause()
	}

	// Validate application to ensure the micro type has been manually set
	if !a.Validate() {
		return nil
	}

	// Get all provisioned devices associated with this application
	if err := a.GetDevices(); err != nil {
		return []error{err}
	}

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

	// Update all online devices associated with this application
	if errs := a.UpdateOnlineDevices(); errs != nil {
		return errs
	}

	// Restart online devices associated with this application
	if err := a.RestartOnlineDevices(); err != nil {
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

	log.WithFields(log.Fields{
		"Pause delay": delay,
	}).Debug("Initialise process")
}

func pause() {
	State.CurrentStatus = PAUSED
	log.WithFields(log.Fields{
		"Status": State.CurrentStatus,
	}).Info("Process status")

	for State.TargetStatus == PAUSED {
		time.Sleep(delay * time.Second)
	}

	State.CurrentStatus = RUNNING
	log.WithFields(log.Fields{
		"Status": State.CurrentStatus,
	}).Info("Process status")
}
