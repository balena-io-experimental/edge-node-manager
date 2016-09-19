package process

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/proxyvisor"
)

// Run processes the application, checking for new commits, provisioning and updating devices
func Run(app *application.Application) []error {
	log.Info("----------------------------------------------------------------------------------------------------")

	if app.Micro == "" || app.Radio == "" {
		return []error{fmt.Errorf("Micro or radio type not set")}
	}

	log.WithFields(log.Fields{
		"Application": app,
	}).Info("Processing application")

	// // Check for latest commit, extract firmware if needed
	// if err := app.ParseCommit(); err != nil {
	// 	return []error{err}
	// }
	// log.WithFields(log.Fields{
	// 	"Directory": app.Directory,
	// 	"Commit":    app.Commit,
	// }).Info("Firmware")

	// Load application devices i.e. already provisioned
	appDevices, err := database.LoadDevices(app.UUID)
	if err != nil {
		return []error{err}
	}
	log.WithFields(log.Fields{
		"Number": len(appDevices),
	}).Info("Application devices")

	if log.GetLevel() == log.DebugLevel {
		for _, appDevice := range appDevices {
			log.WithFields(log.Fields{
				"Device": appDevice,
			}).Debug("Application devices")
		}
	}

	// Get online devices associated with this application
	onlineDevices, err := app.Type.Radio.Scan(app.Name, 10)
	if err != nil {
		return []error{err}
	}
	log.WithFields(log.Fields{
		"Number": len(onlineDevices),
	}).Info("Online devices")

	if log.GetLevel() == log.DebugLevel {
		for localUUID := range onlineDevices {
			log.WithFields(log.Fields{
				"Local UUID": localUUID,
			}).Debug("Online devices")
		}
	}

	//Provision online devices
	for localUUID := range onlineDevices {
		if _, exists := appDevices[localUUID]; exists {
			log.WithFields(log.Fields{
				"Local UUID": localUUID,
			}).Debug("Device already provisioned")
			continue
		}

		log.WithFields(log.Fields{
			"Local UUID": localUUID,
		}).Info("Device not provisioned")

		resinUUID, name, errs := proxyvisor.DependantDeviceProvision(app.UUID)
		if errs != nil {
			return errs
		}

		newDevice := device.New(app.Type, localUUID, resinUUID, name, app.UUID, app.Name, app.Commit)
		newDevice, err = database.SaveDevice(newDevice)
		if err != nil {
			return []error{err}
		}

		appDevices[newDevice.LocalUUID] = newDevice

		log.WithFields(log.Fields{
			"Device": newDevice,
		}).Debug("Device provisioned")
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
			return []error{err}
		} else if online {
			appDevice.State = device.ONLINE
			appDevice.LastSeen = time.Now()

			// if appDevice.Commit == app.Commit {
			// 	continue
			// }

			// if err := appDevice.Cast().Update(app.Commit, app.Directory); err != nil {
			// 	return []error{err}
			// }
		}
	}

	// // Update DB with device changes
	// //TODO: should probably update after every action or at least cache
	// if err := database.UpdateDevices(appDevices); err != nil {
	// 	return []error{err}
	// }

	return nil
}
