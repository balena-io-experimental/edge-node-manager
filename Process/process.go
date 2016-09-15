package process

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/proxyvisor"
)

// Process processes the application, checking for new commits, provisioning and updating devices
func Process(app *application.Application) error {
	log.Info("----------------------------------------------------------------------------------------------------")
	log.WithFields(log.Fields{
		"Application UUID":       app.UUID,
		"Application name":       app.Name,
		"Application micro type": app.Type.Micro,
		"Application radio type": app.Type.Radio,
	}).Info("Processing application")

	// Check for latest commit, extract firmware if needed
	if err := app.ParseCommit(); err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Directory": app.Directory,
		"Commit":    app.Commit,
	}).Info("Firmware")

	// Load application devices i.e. already provisioned
	appDevices, err := database.LoadDevices(app.UUID)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Number": len(appDevices),
	}).Info("Application devices")

	// TODO: research how to loop within the logger so that this loop only runs
	// when the log level is set to debug
	for _, appDevice := range appDevices {
		log.WithFields(log.Fields{
			"Device": appDevice,
		}).Debug("Application devices")
	}

	// Get online devices associated with this application
	onlineDevices, err := app.Type.Radio.Scan(app.Name, 10)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Number": len(onlineDevices),
	}).Info("Online devices")

	// TODO: research how to loop within the logger so that this loop only runs
	// when the log level is set to debug
	for onlineDevice := range onlineDevices {
		log.WithFields(log.Fields{
			"Device": onlineDevice,
		}).Debug("Online devices")
	}

	//Provision online devices
	for onlineDevice := range onlineDevices {
		if _, exists := appDevices[onlineDevice]; exists {
			log.WithFields(log.Fields{
				"Device": onlineDevice,
			}).Debug("Device already provisioned")
			continue
		}

		log.WithFields(log.Fields{
			"Device": onlineDevice,
		}).Info("Device not provisioned")

		resinUUID, err := proxyvisor.NewDevice(app.UUID)
		if err != nil {
			return err
		}
		newDevice, err := app.NewDevice(onlineDevice, resinUUID)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"Device": newDevice,
		}).Debug("Device provisioned")

		appDevices[newDevice.LocalUUID] = newDevice

		log.WithFields(log.Fields{
			"Device": onlineDevice,
		}).Info("Device provisioned")
	}

	// Update device state based on whether the device is online
	// Firmware updates would also be triggered in this section
	for _, appDevice := range appDevices {
		appDevice.State = device.OFFLINE

		_, exists := onlineDevices[appDevice.LocalUUID]
		if !exists {
			continue
		}

		online, err := appDevice.Cast().Online()

		if err != nil {
			return err
		} else if online {
			appDevice.State = device.ONLINE
			appDevice.LastSeen = time.Now()

			if appDevice.Commit == app.Commit {
				continue
			}

			if err := appDevice.Cast().Update(app.Commit, app.Directory); err != nil {
				return err
			}
		}
	}

	// Update DB with device changes
	//TODO: should probably update after every action or at least cache
	return database.UpdateDevices(appDevices)
}
