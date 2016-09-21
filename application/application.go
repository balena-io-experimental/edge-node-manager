package application

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/radio"
	"github.com/josephroberts/edge-node-manager/supervisor"
	tarinator "github.com/verybluebot/tarinator-go"
)

// Uses the tarinator-go package
// https://github.com/verybluebot/tarinator-go

// List holds all the applications assigned to the edge-node-manager
// Key is the applicationUUID
var List map[int]*Application

// Application contains all the variables needed to define an application
type Application struct {
	UUID          int         `json:"appId"`
	Name          string      `json:"name"`
	Commit        string      `json:"-"`      // Ignore this when unmarshalling from the proxyvisor as we want to set the target commit
	TargetCommit  string      `json:"commit"` // Set json tag to commit as the proxyvisor has no concept of target commit
	Env           interface{} `json:"env"`
	DeviceType    string      `json:"device_type"`
	device.Type   `json:"type"`
	Devices       map[string]*device.Device // Key is the device's localUUID
	OnlineDevices map[string]bool           // Key is the device's localUUID
	FilePath      string
}

func (a Application) String() string {
	return fmt.Sprintf(
		"UUID: %d, "+
			"Name: %s, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Env: %v, "+
			"Device type: %s, "+
			"Micro type: %s, "+
			"Radio type: %s",
		a.UUID,
		a.Name,
		a.Commit,
		a.TargetCommit,
		a.Env,
		a.DeviceType,
		a.Type.Micro,
		a.Type.Radio)
}

func init() {
	log.SetLevel(config.GetLogLevel())

	List = make(map[int]*Application)

	bytes, errs := supervisor.DependantApplicationsList()
	if errs != nil {
		log.WithFields(log.Fields{
			"Errors": errs,
		}).Fatal("Unable to get the dependant application list")
	}

	var buffer []Application
	if err := json.Unmarshal(bytes, &buffer); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to unmarshal the dependant application list")
	}

	for key := range buffer {
		UUID := buffer[key].UUID
		List[UUID] = &buffer[key]
	}

	// For now we have to manually initialise an applications micro and radio type
	// This is because the device type returned from the supervisor is always edge
	initApplication(13015, micro.NRF51822, radio.BLUETOOTH)

	log.Debug("Initialised applications")
}

func initApplication(UUID int, micro micro.Type, radio radio.Type) {
	if _, exists := List[UUID]; !exists {
		log.WithFields(log.Fields{
			"UUID": UUID,
		}).Fatal("Application does not exist")
	}

	List[UUID].Type = device.Type{
		Micro: micro,
		Radio: radio,
	}
}

// Validate ensures the application micro and radio type has been manually set
func (a Application) Validate() bool {
	if a.Micro == "" || a.Radio == "" {
		log.WithFields(log.Fields{
			"Application": a,
			"Error":       "Application micro or radio type not set",
		}).Warn("Processing application")
		return false
	}

	log.WithFields(log.Fields{
		"Application": a,
	}).Info("Processing application")

	return true
}

// CheckCommit checks whether there is a new target commit and extracts if necessary
func (a *Application) CheckCommit() error {
	if a.Commit == a.TargetCommit {
		return nil
	}

	a.FilePath = config.GetAssetsDir()
	a.FilePath = path.Join(a.FilePath, strconv.Itoa(a.UUID))
	a.FilePath = path.Join(a.FilePath, a.TargetCommit)
	tarPath := path.Join(a.FilePath, "binary.tar")

	log.WithFields(log.Fields{
		"File path":     a.FilePath,
		"Tar path":      tarPath,
		"Target commit": a.TargetCommit,
	}).Debug("Application firmware")

	if err := tarinator.UnTarinate(a.FilePath, tarPath); err != nil {
		return err
	}

	a.Commit = a.TargetCommit

	log.WithFields(log.Fields{
		"File path":     a.FilePath,
		"Tar path":      tarPath,
		"Target commit": a.TargetCommit,
	}).Info("Application firmware extracted")

	return nil
}

// GetDevices gets all provisioned devices associated with the application
func (a *Application) GetDevices() error {
	var err error
	if a.Devices, err = device.GetAll(a.UUID); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Number": len(a.Devices),
	}).Info("Application devices")

	if log.GetLevel() == log.DebugLevel {
		for _, value := range a.Devices {
			log.WithFields(log.Fields{
				"Device": value,
			}).Debug("Application device")
		}
	}

	return nil
}

// PutDevices puts all provisioned devices associated with the application
func (a *Application) PutDevices() error {
	return device.PutAll(a.UUID, a.Devices)
}

// GetOnlineDevices gets all online devices associated with the application
func (a *Application) GetOnlineDevices() error {
	// Scan for devices with an ID that matches the applicationUUID
	var err error
	if a.OnlineDevices, err = a.Type.Radio.Scan(strconv.Itoa(a.UUID), 10); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Number": len(a.OnlineDevices),
	}).Info("Application online devices")

	if log.GetLevel() == log.DebugLevel {
		for key := range a.OnlineDevices {
			log.WithFields(log.Fields{
				"Local UUID": key,
			}).Debug("Online device")
		}
	}

	return nil
}

// ProvisionDevices provisions all non-provisoned online devices associated with the application
func (a *Application) ProvisionDevices() []error {
	for key := range a.OnlineDevices {
		// Check if the online device localUUID is present in devices
		// If it is then the device is already provisioned
		if _, exists := a.Devices[key]; exists {
			log.WithFields(log.Fields{
				"Local UUID": key,
			}).Debug("Device already provisioned")
			continue
		}

		log.WithFields(log.Fields{
			"Local UUID": key,
		}).Info("Device not provisioned")

		deviceUUID, deviceName, errs := supervisor.DependantDeviceProvision(a.UUID)
		if errs != nil {
			return errs
		}

		err := device.New(a.Type, key, deviceUUID, deviceName, a.UUID, a.Name, a.Commit)
		if err != nil {
			return []error{err}
		}

		newDevice, err := device.Get(a.UUID, deviceUUID)
		if err != nil {
			return []error{err}
		}

		a.Devices[newDevice.LocalUUID] = newDevice

		log.WithFields(log.Fields{
			"Device": newDevice,
		}).Info("Device provisioned")
	}

	return nil
}

// SetState sets the state for all provisioned devices associated with the application
func (a *Application) SetState(state device.State) {
	for _, d := range a.Devices {
		d.SetState(state)
	}
}

// UpdateOnlineDevices updates all online devices associated with the application
// State and last time seen fields
// Firmware if a new commit is available
func (a *Application) UpdateOnlineDevices() error {
	for key := range a.OnlineDevices {
		d := a.Devices[key]

		online, err := d.Online()
		if err != nil {
			return err
		}

		if online {
			d.SetState(device.ONLINE)

			// Get the target commit as it may have been set by the supervisor since we loaded all the application devices
			bytes, err := database.GetDeviceField(a.UUID, d.UUID, "targetCommit")
			if err != nil {
				return err
			}
			d.TargetCommit = (string)(bytes)

			if d.Commit == d.TargetCommit {
				log.WithFields(log.Fields{
					"Device": d,
				}).Debug("Device upto date")
				continue
			}

			log.WithFields(log.Fields{
				"Device": d,
			}).Info("Device not upto date")

			if err := d.Update(a.FilePath); err != nil {
				return err
			}
			d.Commit = d.TargetCommit

			log.WithFields(log.Fields{
				"Device": d,
			}).Info("Device updated")
		}
	}

	return nil
}
