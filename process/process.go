package process

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/application"
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

	// In app check if target matches commit, extract if not
	app.Commit = app.TargetCommit

	// // Check for latest commit, extract firmware if needed
	// if err := app.ParseCommit(); err != nil {
	// 	return []error{err}
	// }
	// log.WithFields(log.Fields{
	// 	"Directory": app.Directory,
	// 	"Commit":    app.Commit,
	// }).Info("Firmware")

	// Load application devices i.e. already provisioned
	applicationDevices, err := device.GetAll(app.UUID)
	if err != nil {
		return []error{err}
	}
	log.WithFields(log.Fields{
		"Number": len(applicationDevices),
	}).Info("Application devices")

	if log.GetLevel() == log.DebugLevel {
		for _, value := range applicationDevices {
			log.WithFields(log.Fields{
				"Device": value,
			}).Debug("Application device")
		}
	}

	// Get online devices associated with this application
	onlineDevices, err := app.Type.Radio.Scan(app.Name, 10) //TODO change this to uuid
	if err != nil {
		return []error{err}
	}
	log.WithFields(log.Fields{
		"Number": len(onlineDevices),
	}).Info("Online devices")

	if log.GetLevel() == log.DebugLevel {
		for key := range onlineDevices {
			log.WithFields(log.Fields{
				"Local UUID": key,
			}).Debug("Online device")
		}
	}

	//Provision online devices
	for key := range onlineDevices {
		if _, exists := applicationDevices[key]; exists {
			log.WithFields(log.Fields{
				"Local UUID": key,
			}).Debug("Device already provisioned")
			continue
		}

		log.WithFields(log.Fields{
			"Local UUID": key,
		}).Info("Device not provisioned")

		deviceUUID, name, errs := proxyvisor.DependantDeviceProvision(app.UUID)
		if errs != nil {
			return errs
		}

		err := device.New(app.Type, key, deviceUUID, name, app.UUID, app.Name, app.Commit)
		if err != nil {
			return []error{err}
		}

		// newDevice, err := device.Get(app.UUID, deviceUUID)
		// if err != nil {
		// 	return []error{err}
		// }

		// applicationDevices[newDevice.LocalUUID] = newDevice

		// log.WithFields(log.Fields{
		// 	"Device": newDevice,
		// }).Debug("Device provisioned")
	}

	// // Update device state based on whether the device is online
	// // Firmware updates would also be triggered in this section
	// for _, appDevice := range appDevices {
	// 	appDevice.State = device.OFFLINE

	// 	_, exists := onlineDevices[appDevice.LocalUUID]
	// 	if !exists {
	// 		continue
	// 	}

	// 	online, err := appDevice.Cast().Online()

	// 	if err != nil {
	// 		return []error{err}
	// 	} else if online {
	// 		appDevice.State = device.ONLINE
	// 		appDevice.LastSeen = time.Now()

	// 		// if appDevice.Commit == app.Commit {
	// 		// 	continue
	// 		// }

	// 		// if err := appDevice.Cast().Update(app.Commit, app.Directory); err != nil {
	// 		// 	return []error{err}
	// 		// }
	// 	}
	// }

	// // Update DB with device changes
	// //TODO: should probably update after every action or at least cache
	// if err := database.UpdateDevices(appDevices); err != nil {
	// 	return []error{err}
	// }

	return nil
}
