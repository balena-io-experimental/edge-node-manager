package application

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/board"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/device/status"
	"github.com/josephroberts/edge-node-manager/supervisor"
	tarinator "github.com/verybluebot/tarinator-go"
)

// List holds all the applications assigned to the edge-node-manager
// Key is the applicationUUID
var List map[int]*Application

type Application struct {
	ResinUUID     int    `json:"id"`
	Name          string `json:"name"`
	BoardType     board.Type
	Commit        string      `json:"-"`      // Ignore this when unmarshalling from the supervisor as we want to set the target commit
	TargetCommit  string      `json:"commit"` // Set json tag to commit as the supervisor has no concept of target commit
	Config        interface{} `json:"config"`
	FilePath      string
	Devices       map[string]*device.Device // Key is the device's localUUID
	OnlineDevices map[string]bool           // Key is the device's localUUID
}

func (a Application) String() string {
	return fmt.Sprintf(
		"UUID: %d, "+
			"Name: %s, "+
			"Board type: %s, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Config: %v, "+
			"File path: %s",
		a.ResinUUID,
		a.Name,
		a.BoardType,
		a.Commit,
		a.TargetCommit,
		a.Config,
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

	for a := range buffer {
		ResinUUID := buffer[a].ResinUUID
		List[ResinUUID] = &buffer[a]

		log.WithFields(log.Fields{
			"Application": List[ResinUUID],
		}).Debug("Dependant application")
	}

	initApplication(14495, board.NRF51822DK)

	for _, a := range List {
		if err := a.getDevices(); err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("Unable to load application devices")
		}
	}

	log.Debug("Initialised applications")
}

func (a Application) Validate() bool {
	if a.BoardType == "" {
		log.WithFields(log.Fields{
			"Application": a,
			"Error":       "Application board type not set",
		}).Warn("Processing application")
		return false
	}

	log.WithFields(log.Fields{
		"Application": a,
	}).Info("Processing application")

	if log.GetLevel() == log.DebugLevel {
		for _, d := range a.Devices {
			log.WithFields(log.Fields{
				"Device": d,
			}).Debug("Application device")
		}
	}

	return true
}

func (a *Application) GetOnlineDevices() error {
	board := board.Create(a.BoardType, "")

	var err error
	if a.OnlineDevices, err = board.Scan(a.ResinUUID); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Number": len(a.OnlineDevices),
	}).Info("Application online devices")

	if log.GetLevel() != log.DebugLevel {
		return nil
	}

	for localUUID := range a.OnlineDevices {
		log.WithFields(log.Fields{
			"Local UUID": localUUID,
		}).Debug("Online device")
	}

	return nil
}

func (a *Application) ProvisionDevices() []error {
	for localUUID := range a.OnlineDevices {
		if _, exists := a.Devices[localUUID]; exists {
			log.WithFields(log.Fields{
				"Local UUID": localUUID,
			}).Debug("Device already provisioned")
			continue
		}

		log.WithFields(log.Fields{
			"Local UUID": localUUID,
		}).Info("Device not provisioned")

		resinUUID, name, config, env, errs := supervisor.DependantDeviceProvision(a.ResinUUID)
		if errs != nil {
			return errs
		}

		device := device.Create(a.BoardType, name, localUUID, resinUUID, a.ResinUUID, a.Name, a.Commit, config, env)

		a.Devices[device.LocalUUID] = device

		if err := a.putDevice(device.LocalUUID); err != nil {
			return []error{err}
		}

		log.WithFields(log.Fields{
			"Name": device.Name,
		}).Info("Device provisioned")
	}

	return nil
}

func (a *Application) SetOfflineDeviceStatus() []error {
	for _, d := range a.Devices {
		if _, exists := a.OnlineDevices[d.LocalUUID]; !exists {
			if errs := d.SetStatus(status.OFFLINE); errs != nil {
				return errs
			}
		}
	}

	return nil
}

func (a *Application) UpdateOnlineDevices() []error {
	for localUUID := range a.OnlineDevices {
		d := a.Devices[localUUID]

		online, err := d.Board.Online()
		if err != nil {
			return []error{err}
		}

		if !online {
			d.SetStatus(status.OFFLINE)
			return nil
		}

		d.SetStatus(status.IDLE)

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

		d.SetStatus(status.INSTALLING)
		if err := d.Board.Update(a.FilePath); err != nil {
			return []error{err}
		}
		d.Commit = d.TargetCommit
		d.SetStatus(status.IDLE)

		log.WithFields(log.Fields{
			"Name": d.Name,
		}).Info("Finished update")
	}

	return nil
}

func (a *Application) RestartOnlineDevices() error {
	for localUUID := range a.OnlineDevices {
		d := a.Devices[localUUID]

		if !d.RestartFlag {
			continue
		}

		online, err := d.Board.Online()
		if err != nil {
			return err
		}

		if !online {
			d.SetStatus(status.OFFLINE)
			return nil
		}

		d.SetStatus(status.IDLE)

		if err = d.Board.Restart(); err != nil {
			return err
		}

		d.RestartFlag = false

		log.WithFields(log.Fields{
			"Device": d,
		}).Info("Device restarted")
	}

	return nil
}

func initApplication(UUID int, boardType board.Type) {
	if _, exists := List[UUID]; !exists {
		log.WithFields(log.Fields{
			"UUID": UUID,
		}).Fatal("Application does not exist")
	}

	List[UUID].BoardType = boardType
}

func (a *Application) PutDevices() error {
	buffer := make(map[string][]byte)
	for _, d := range a.Devices {
		bytes, err := json.Marshal(d)
		if err != nil {
			return err
		}
		buffer[d.ResinUUID] = bytes
	}

	return database.PutDevices(a.ResinUUID, buffer)
}

func (a *Application) getDevices() error {
	a.Devices = make(map[string]*device.Device)

	buffer, err := database.GetDevices(a.ResinUUID)
	if err != nil {
		return err
	}

	for _, d := range buffer {
		var device device.Device
		if err = json.Unmarshal(d, &device); err != nil {
			return err
		}

		a.Devices[device.LocalUUID] = &device
	}

	log.WithFields(log.Fields{
		"Number": len(a.Devices),
	}).Info("Application devices")

	return nil
}

func (a *Application) putDevice(localUUID string) error {
	d := a.Devices[localUUID]
	bytes, err := json.Marshal(d)
	if err != nil {
		return err
	}

	return database.PutDevice(a.ResinUUID, d.LocalUUID, d.ResinUUID, bytes)
}

func (a *Application) checkCommit() error {
	if a.Commit == a.TargetCommit {
		return nil
	}

	if err := supervisor.DependantApplicationUpdate(a.ResinUUID, a.TargetCommit); err != nil {
		return err
	}

	a.FilePath = config.GetAssetsDir()
	a.FilePath = path.Join(a.FilePath, strconv.Itoa(a.ResinUUID))
	a.FilePath = path.Join(a.FilePath, a.TargetCommit)
	tarPath := path.Join(a.FilePath, "binary.tar")

	if err := tarinator.UnTarinate(a.FilePath, tarPath); err != nil {
		return err
	}

	a.Commit = a.TargetCommit

	log.Info("Application firmware extracted")

	return nil
}
