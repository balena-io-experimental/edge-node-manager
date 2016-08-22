package application

import (
	"encoding/json"
	"log"
	"time"

	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/helper"
	"github.com/josephroberts/edge-node-manager/proxyvisor"
	"github.com/josephroberts/edge-node-manager/radio"
)

type ApplicationInterface interface {
	Process() error
	loadDevices() (map[string]device.DeviceInterface, error)
	updateDevices(devices map[string]device.DeviceInterface) error
	createDevice(localUUID string) (device.DeviceInterface, string, error)
}

type Application struct {
	Name   string
	Device device.SupportedDevice
	Radio  radio.RadioInterface
}

func (a *Application) Process() error {
	log.Printf("-----------------------------------------------------------------\r\n")
	log.Printf("Application name: %s\r\n", a.Name)
	log.Printf("Application device type: %s\r\n", a.Device)
	log.Printf("Application radio type: %T\r\n", a.Radio)

	application, commit, err := helper.GetApplication(config.Persistant, a.Name)
	if err != nil {
		log.Printf("No application location or commit found\r\n")
	} else {
		log.Printf("Application location: %s\r\n", application)
		log.Printf("Application commit: %s\r\n", commit)
	}

	applicationDevices, _ := a.loadDevices()
	log.Printf("%d application devices found\r\n", len(applicationDevices))
	for key, applicationDevice := range applicationDevices {
		log.Printf("Id: %s, %s\r\n", key, applicationDevice)
	}

	onlineDevices, err := a.Radio.Scan(a.Name, 10)
	log.Printf("%d online devices found\r\n", len(onlineDevices))
	if err != nil {
		log.Fatal("Failed to scan for online devices")
	} else {
		for _, onlineDevice := range onlineDevices {
			log.Printf("%s\r\n", onlineDevice)
		}
	}

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
			device, key, err := a.createDevice(onlineDevice)
			if err != nil {
				log.Fatal("Failed to scan for online devices")
			}
			applicationDevices[key] = device

		} else {
			log.Printf("Device exists: %s\r\n", onlineDevice)
		}
	}

	for _, applicationDevice := range applicationDevices {
		online, err := a.Radio.Online(applicationDevice.GetDevice().LocalUUID, 10)
		if err != nil {
			log.Fatal("Failed to scan for online devices")
		}

		if online {
			applicationDevice.GetDevice().State = device.ONLINE
			applicationDevice.GetDevice().LastSeen = time.Now()
			//applicationDevice.Update(application, commit)
		} else {
			applicationDevice.GetDevice().State = device.OFFLINE
		}

	}

	a.updateDevices(applicationDevices)

	return nil
}

func (a *Application) loadDevices() (map[string]device.DeviceInterface, error) {
	resp, err := database.Query("applicationUUID", a.Name)
	if err != nil {
		log.Fatal(err)
	}

	buffer := make(map[string]interface{})
	err = json.Unmarshal(resp, &buffer)

	devices := make(map[string]device.DeviceInterface)

	for key, value := range buffer {
		b, _ := json.Marshal(value)
		i := device.Create(a.Device)
		i.GetDevice().Deserialise(b)
		i.GetDevice().Radio = a.Radio
		devices[key] = i
	}

	return devices, err
}

func (a *Application) updateDevices(devices map[string]device.DeviceInterface) error {
	for key, value := range devices {
		buffer, err := value.Serialise()
		if err != nil {
			log.Fatal(err)
		}

		err = database.Update(key, buffer)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (a *Application) createDevice(localUUID string) (device.DeviceInterface, string, error) {
	i := device.Create(a.Device)
	d := i.GetDevice()
	d.LocalUUID = localUUID
	d.ApplicationUUID = a.Name
	resinUUID, err := proxyvisor.Provision()
	if err != nil {
		log.Fatal("Failed to provision device")
	}
	d.ResinUUID = resinUUID
	d.Commit = ""
	d.LastSeen = time.Now()
	d.State = device.ONLINE
	d.Progress = 0.0
	d.Radio = a.Radio

	b, err := d.Serialise()
	if err != nil {
		log.Fatal("Failed to provision device")
	}

	b, err = database.Insert(b)
	if err != nil {
		log.Fatal("Failed to provision device")
	}

	return i, string(b), nil
}
