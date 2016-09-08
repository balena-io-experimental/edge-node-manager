package application

import (
	"errors"
	"io/ioutil"
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
		"Directory": a.Directory,
		"Commit":    a.Commit,
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
	for onlineDevice, _ := range onlineDevices {
		log.WithFields(log.Fields{
			"Device": onlineDevice,
		}).Debug("Online devices")
	}

	//Provision online devices
	for onlineDevice, _ := range onlineDevices {
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

		online, err := appDevice.ChooseType().Online()

		if err != nil {
			return err
		} else if online {
			appDevice.State = device.ONLINE
			appDevice.LastSeen = time.Now()
		}
	}

	// Update DB with device changes
	if err := database.UpdateDevices(appDevices); err != nil {
		return err
	}

	return nil
}

func (a *Application) parseCommit() error {
	if a.Directory == "" {
		a.Directory = filepath.Join(config.GetPersistantDirectory(), a.UUID)
	}

	commitDirectories, err := ioutil.ReadDir(a.Directory)
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

	tarDirectory := filepath.Join(a.Directory, a.Commit)
	tarPath := filepath.Join(tarDirectory, "binary.tar")
	if err := tarinator.UnTarinate(tarDirectory, tarPath); err != nil {
		return err
	}

	return nil
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
