package application

import (
	"errors"
	"io/ioutil"
	"path"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/firmware"
	"github.com/josephroberts/edge-node-manager/proxyvisor"
	"github.com/verybluebot/tarinator-go"
)

type Application struct {
	UUID string
	device.DeviceType
	firmware.Firmware
}

func (a *Application) Process() error {
	log.Info("----------------------------------------------------------------------------------------------------")
	log.WithFields(log.Fields{
		"Application UUID":       a.UUID,
		"Application micro type": a.DeviceType.Micro,
		"Application radio type": a.DeviceType.Radio,
	}).Info("Processing application")

	// Check for latest commit, extract firmware if needed
	if err := a.parseCommit(); err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Firmware directory": a.Dir,
		"Commit":             a.Commit,
	}).Info("Firmware")

	// Load application devices i.e. already provisioned
	appDevices, err := database.LoadDevices(a.UUID)
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
	onlineDevices, err := a.DeviceType.Radio.Scan(a.UUID, 10)
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

		newDevice, err := a.newDevice(onlineDevice)
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

			if appDevice.Commit == a.Commit {
				continue
			}

			if err := appDevice.Cast().Update(a.Firmware); err != nil {
				return err
			}
		}
	}

	// Update DB with device changes
	//TODO: should probably update after every action or at least cache
	return database.UpdateDevices(appDevices)
}

func (a *Application) parseCommit() error {
	appDir := filepath.Join(config.GetPersistantDirectory(), a.UUID)

	commitDirectories, err := ioutil.ReadDir(appDir)
	if err != nil {
		return err
	} else if len(commitDirectories) == 0 {
		return nil
	} else if len(commitDirectories) > 1 {
		return errors.New("More than one commit found")
	}

	commit := commitDirectories[0].Name()
	if a.Commit == commit {
		return nil
	}

	a.Commit = commit
	a.Dir = path.Join(appDir, a.Commit)

	tarPath := filepath.Join(a.Dir, "binary.tar")
	return tarinator.UnTarinate(a.Dir, tarPath)
}

func (a *Application) newDevice(onlineDevice string) (*device.Device, error) {
	resinUUID, err := proxyvisor.NewDevice(a.UUID)
	if err != nil {
		return &device.Device{}, err
	}

	newDevice := &device.Device{
		DeviceType:      a.DeviceType,
		LocalUUID:       onlineDevice,
		ApplicationUUID: a.UUID,
		ResinUUID:       resinUUID,
		LastSeen:        time.Now(),
		State:           device.ONLINE,
	}

	return database.SaveDevice(newDevice)
}
