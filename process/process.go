package process

import (
	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/device"
)

// Run processes the application, checking for new commits, provisioning and updating devices
func Run(a *application.Application) []error {
	log.Info("----------------------------------------------------------------------------------------------------")

	// Validate application to ensure the micro and radio type has been manually set
	if !a.Validate() {
		return nil
	}

	// Check whether there is a new target commit and extract if necessary
	if err := a.CheckCommit(); err != nil {
		return []error{err}
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

	// Set all provisioned devices associated with this application to OFFLINE
	a.SetState(device.OFFLINE)

	// Update all online devices associated with this application
	// State and last time seen fields
	// Firmware if a new commit is available
	if err := a.UpdateOnlineDevices(); err != nil {
		return []error{err}
	}

	// Put all provisioned devices associated with this application
	if err := a.PutDevices(); err != nil {
		return []error{err}
	}

	return nil
}
