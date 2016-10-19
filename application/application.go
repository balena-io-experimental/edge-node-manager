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
	UUID          int         `json:"id"`
	Name          string      `json:"name"`
	Commit        string      `json:"-"`      // Ignore this when unmarshalling from the supervisor as we want to set the target commit
	TargetCommit  string      `json:"commit"` // Set json tag to commit as the supervisor has no concept of target commit
	Config        interface{} `json:"config"`
	device.Type   `json:"type"`
	FilePath      string
	Devices       map[string]*device.Device // Key is the device's localUUID
	OnlineDevices map[string]bool           // Key is the device's localUUID
}

func (a Application) String() string {
	return fmt.Sprintf(
		"UUID: %d, "+
			"Name: %s, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Config: %v, "+
			"Micro type: %s, "+
			"File path: %s",
		a.UUID,
		a.Name,
		a.Commit,
		a.TargetCommit,
		a.Config,
		a.Type.Micro,
		a.FilePath)
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
			"Data":  ((string)(bytes)),
		}).Fatal("Unable to unmarshal the dependant application list")
	}

	for key := range buffer {
		UUID := buffer[key].UUID
		List[UUID] = &buffer[key]

		log.WithFields(log.Fields{
			"Key":         UUID,
			"Application": List[UUID],
		}).Debug("Dependant application")
	}

	// For now we have to manually initialise an applications micro type
	// This is because the device type returned from the supervisor is always edge
	// initApplication(13829, micro.NRF51822)
	initApplication(14323, micro.MICROBIT)

	log.Debug("Initialised applications")
}

func initApplication(UUID int, micro micro.Type) {
	if _, exists := List[UUID]; !exists {
		log.WithFields(log.Fields{
			"UUID": UUID,
		}).Fatal("Application does not exist")
	}

	List[UUID].Type = device.Type{
		Micro: micro,
	}
}

// Validate ensures the application micro type has been manually set
func (a Application) Validate() bool {
	if a.Micro == "" {
		log.WithFields(log.Fields{
			"Application": a,
			"Error":       "Application micro type not set",
		}).Warn("Processing application")
		return false
	}

	log.WithFields(log.Fields{
		"UUID": a.UUID,
	}).Info("Processing application")

	return true
}

// GetDevices gets all provisioned devices associated with the application
func (a *Application) GetDevices() error {
	a.Devices = make(map[string]*device.Device)

	buffer, err := database.GetDevices(a.UUID)
	if err != nil {
		return err
	}

	for _, value := range buffer {
		var device device.Device
		if err = json.Unmarshal(value, &device); err != nil {
			return err
		}

		a.Devices[device.LocalUUID] = &device
	}

	log.WithFields(log.Fields{
		"Number": len(a.Devices),
	}).Info("Application devices")

	if log.GetLevel() != log.DebugLevel {
		return nil
	}

	for _, value := range a.Devices {
		log.WithFields(log.Fields{
			"Device": value,
		}).Debug("Application device")
	}

	return nil
}

// PutDevices puts all provisioned devices associated with the application
func (a *Application) PutDevices() error {
	buffer := make(map[string][]byte)
	for _, value := range a.Devices {
		bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		buffer[value.UUID] = bytes
	}

	return database.PutDevices(a.UUID, buffer)
}

// GetOnlineDevices gets all online devices associated with the application
func (a *Application) GetOnlineDevices() error {
	// Temporary device used to scan
	tempDevice := &device.Device{
		Type:            a.Type,
		ApplicationUUID: a.UUID,
	}

	var err error
	if a.OnlineDevices, err = tempDevice.Scan(); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Number": len(a.OnlineDevices),
	}).Info("Application online devices")

	if log.GetLevel() != log.DebugLevel {
		return nil
	}

	for key := range a.OnlineDevices {
		log.WithFields(log.Fields{
			"Local UUID": key,
		}).Debug("Online device")
	}

	return nil
}

// ProvisionDevices provisions all non-provisoned online devices associated with the application
func (a *Application) ProvisionDevices() []error {
	for key := range a.OnlineDevices {
		if _, exists := a.Devices[key]; exists {
			log.WithFields(log.Fields{
				"Local UUID": key,
			}).Debug("Device already provisioned")
			continue
		}

		log.WithFields(log.Fields{
			"Local UUID": key,
		}).Info("Device not provisioned")

		deviceUUID, deviceName, deviceConfig, deviceEnv, errs := supervisor.DependantDeviceProvision(a.UUID)
		if errs != nil {
			return errs
		}

		err := device.New(a.Type, key, deviceUUID, deviceName, a.UUID, a.Name, a.Commit, deviceConfig, deviceEnv)
		if err != nil {
			return []error{err}
		}

		newDevice, err := device.Get(a.UUID, deviceUUID)
		if err != nil {
			return []error{err}
		}

		a.Devices[newDevice.LocalUUID] = newDevice

		log.WithFields(log.Fields{
			"Name": newDevice.Name,
		}).Info("Device provisioned")
	}

	return nil
}

// SetOfflineDeviceStatus sets the status for all offline provisioned devices associated with the application
func (a *Application) SetOfflineDeviceStatus() []error {
	for key, d := range a.Devices {
		if _, exists := a.OnlineDevices[key]; !exists {
			if errs := d.SetStatus(device.OFFLINE); errs != nil {
				return errs
			}
		}
	}

	return nil
}

// UpdateOnlineDevices updates all online devices associated with the application
func (a *Application) UpdateOnlineDevices() []error {
	for key := range a.OnlineDevices {
		d := a.Devices[key]

		online, err := d.Online()
		if err != nil {
			return []error{err}
		}

		if !online {
			d.SetStatus(device.OFFLINE)
			return nil
		}

		d.SetStatus(device.IDLE)

		if d.Commit == d.TargetCommit {
			log.WithFields(log.Fields{
				"Device": d,
			}).Debug("Device up to date")
			continue
		}

		log.WithFields(log.Fields{
			"Name": d.Name,
		}).Info("Device not up to date")

		if err := a.checkCommit(); err != nil {
			return []error{err}
		}

		log.WithFields(log.Fields{
			"Name": d.Name,
		}).Info("Starting update")

		d.SetStatus(device.INSTALLING)
		if err := d.Update(a.FilePath); err != nil {
			return []error{err}
		}
		d.Commit = d.TargetCommit
		d.SetStatus(device.IDLE)

		log.WithFields(log.Fields{
			"Name": d.Name,
		}).Info("Finished update")
	}

	return nil
}

// RestartOnlineDevices restarts online devices associated with the application if the restart flag is set
func (a *Application) RestartOnlineDevices() error {
	for key := range a.OnlineDevices {
		d := a.Devices[key]

		if !d.RestartFlag {
			continue
		}

		online, err := d.Online()
		if err != nil {
			return err
		}

		if !online {
			d.SetStatus(device.OFFLINE)
			return nil
		}

		d.SetStatus(device.IDLE)

		if err = d.Restart(); err != nil {
			return err
		}

		d.RestartFlag = false

		log.WithFields(log.Fields{
			"Device": d,
		}).Info("Device restarted")
	}

	return nil
}

func (a *Application) checkCommit() error {
	if a.Commit == a.TargetCommit {
		return nil
	}

	if err := supervisor.DependantApplicationUpdate(a.UUID, a.TargetCommit); err != nil {
		return err
	}

	a.FilePath = config.GetAssetsDir()
	a.FilePath = path.Join(a.FilePath, strconv.Itoa(a.UUID))
	a.FilePath = path.Join(a.FilePath, a.TargetCommit)
	tarPath := path.Join(a.FilePath, "binary.tar")

	if err := tarinator.UnTarinate(a.FilePath, tarPath); err != nil {
		return err
	}

	a.Commit = a.TargetCommit

	log.Info("Application firmware extracted")

	return nil
}
