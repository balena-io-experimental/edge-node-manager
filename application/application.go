package application

import (
	"fmt"
	"log"
	"time"

	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/helper"
	"github.com/josephroberts/edge-node-manager/proxyvisor"
)

type Application struct {
	UUID       string
	DeviceType *device.Type
}

func (a *Application) Process() error {
	log.Printf("-----------------------------------------------------------------\r\n")
	log.Printf("Application UUID: %s\r\n", a.UUID)
	log.Printf("Application device type: %s\r\n", a.DeviceType.Device)
	log.Printf("Application radio type: %s\r\n", a.DeviceType.Radio)

	// Get latest firmware .zip and commit
	application, commit, err := helper.GetApplication(config.GetPersistantDirectory(), a.UUID)
	if err != nil {
		log.Printf("No application location or commit found\r\n")
	} else {
		log.Printf("Application location: %s\r\n", application)
		log.Printf("Application commit: %s\r\n", commit)
	}

	// Load applicaion devices i.e. already provisioned
	applicationDevices, err := database.LoadDevices(a.UUID, a.DeviceType)
	if err != nil {
		return fmt.Errorf("Unable to load devices: %v", err)
	}
	log.Printf("Application devices found: %d\r\n", len(applicationDevices))
	for key, device := range applicationDevices {
		log.Printf("Key: %d, %s\r\n", key, device)
	}

	// Get online devices associated with this application
	onlineDevices, err := a.DeviceType.Scan(a.UUID, 10)
	if err != nil {
		return fmt.Errorf("Unable to scan for online devices: %v", err)
	}
	log.Printf("Online devices found: %d\r\n", len(onlineDevices))
	for _, device := range onlineDevices {
		log.Printf("Device: %s\r\n", device)
	}

	// Provision online devices if not already provisioned
	for _, onlineDevice := range onlineDevices {
		exists := false
		for _, applicationDevice := range applicationDevices {
			if applicationDevice.GetDevice().LocalUUID == onlineDevice {
				exists = true
				break
			}
		}
		if !exists {
			log.Printf("Provisioning device: %s\r\n", onlineDevice)
			if resinUUID, err := proxyvisor.Provision(); err != nil {
				return fmt.Errorf("Unable to provision online device: %v", err)
			} else {
				provisionedDevice := device.Provision(a.DeviceType, onlineDevice, a.UUID, resinUUID)
				if key, err := database.SaveDevice(provisionedDevice); err != nil {
					return fmt.Errorf("Unable to save provisioned device: %v", err)
				} else {
					applicationDevices[key] = provisionedDevice
				}
			}
		} else {
			log.Printf("Device exists: %s\r\n", onlineDevice)
		}
	}

	// Update device state based on whether the device is online
	// Firmware updates would also be triggered in this section
	for _, applicationDevice := range applicationDevices {
		applicationDevice.GetDevice().State = device.OFFLINE
		for _, onlineDevice := range onlineDevices {
			if applicationDevice.GetDevice().LocalUUID == onlineDevice {
				if online, err := applicationDevice.Online(); err != nil {
					return fmt.Errorf("Unable to check if device online: %v", err)
				} else if online {
					applicationDevice.GetDevice().State = device.ONLINE
					applicationDevice.GetDevice().LastSeen = time.Now()
				}
			}
		}
	}

	// Update DB with device changes
	database.UpdateDevices(applicationDevices)

	return nil
}
