package process

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/application"
)

var (
	PauseTarget = false
	PauseState  = false
)

// Run processes the application, checking for new commits, provisioning and updating devices
func Run(a *application.Application, delay time.Duration) []error {
	// Pause the process if necessary
	for PauseTarget {
		PauseState = true
		time.Sleep(delay * time.Second)
	}
	PauseState = false

	log.Info("----------------------------------------------------------------------------------------------------")

	// Validate application to ensure the micro and radio type has been manually set
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
