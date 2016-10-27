package application

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/board"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/database"
	"github.com/resin-io/edge-node-manager/device"
	"github.com/resin-io/edge-node-manager/device/status"
	"github.com/resin-io/edge-node-manager/supervisor"
	tarinator "github.com/verybluebot/tarinator-go"
)

// List holds all the applications assigned to the edge-node-manager
// Key is the applicationUUID
var List map[int]*Application

type Application struct {
	ResinUUID     int                       `json:"id"`
	Name          string                    `json:"name"`
	BoardType     board.Type                `json:"-"`
	Commit        string                    `json:"-"`      // Ignore this when unmarshalling from the supervisor as we want to set the target commit
	TargetCommit  string                    `json:"commit"` // Set json tag to commit as the supervisor has no concept of target commit
	Config        map[string]interface{}    `json:"config"`
	FilePath      string                    `json:"-"`
	Devices       map[string]*device.Device `json:"-"` // Key is the device's localUUID
	OnlineDevices map[string]bool           `json:"-"` // Key is the device's localUUID
	deleteFlag    bool
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

func Load() []error {
	bytes, errs := supervisor.DependantApplicationsList()
	if errs != nil {
		return errs
	}

	var buffer []Application
	if err := json.Unmarshal(bytes, &buffer); err != nil {
		return []error{err}
	}

	for _, application := range List {
		application.deleteFlag = true
	}

	for key := range buffer {
		ResinUUID := buffer[key].ResinUUID

		if application, exists := List[ResinUUID]; exists {
			application.deleteFlag = false
			continue
		}

		List[ResinUUID] = &buffer[key]
		application := List[ResinUUID]

		// Start temporary
		if ResinUUID == 14539 {
			application.Config["BOARD"] = "micro:bit"
		}
		if ResinUUID == 14495 {
			application.Config["BOARD"] = "nRF51822-DK"
		}
		// End temporary

		if _, exists := application.Config["BOARD"]; exists {
			application.BoardType = (board.Type)(application.Config["BOARD"].(string))
		}

		if err := application.GetDevices(); err != nil {
			return []error{err}
		}
	}

	for key, application := range List {
		if application.deleteFlag {
			delete(List, key)
		}
	}

	return nil
}

func (a *Application) GetOnlineDevices() error {
	board, err := board.Create(a.BoardType, "", nil)
	if err != nil {
		return err
	}

	if a.OnlineDevices, err = board.Scan(a.ResinUUID); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Number of online devices": len(a.OnlineDevices),
	}).Info("Processing application")

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
		}).Info("Provisioning device")

		resinUUID, name, config, env, errs := supervisor.DependantDeviceProvision(a.ResinUUID)
		if errs != nil {
			return errs
		}

		d, err := device.Create(a.BoardType, name, localUUID, resinUUID, a.ResinUUID, a.Name, a.Commit, config, env)
		if err != nil {
			return []error{err}
		}

		a.Devices[d.LocalUUID] = d

		log.WithFields(log.Fields{
			"Name":       d.Name,
			"Local UUID": localUUID,
		}).Info("Provisioned device")
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
			"Device": d,
		}).Debug("Device not up to date")

		if err := a.checkCommit(); err != nil {
			return []error{err}
		}

		log.WithFields(log.Fields{
			"Name": d.Name,
		}).Info("Starting update")

		d.SetStatus(status.INSTALLING)

		if err := d.Board.Update(a.FilePath); err != nil {
			log.WithFields(log.Fields{
				"Name": d.Name,
			}).Error("Update failed")
			d.SetStatus(status.IDLE)
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

func (a *Application) HandleFlags() error {
	if err := a.handleDeleteFlag(); err != nil {
		return err
	}

	if err := a.handleRestartFlag(); err != nil {
		return err
	}

	return nil
}

func (a *Application) PutDevices() error {
	buffer := make(map[string][]byte)
	for _, d := range a.Devices {
		bytes, err := d.Marshall()
		if err != nil {
			return err
		}
		buffer[d.ResinUUID] = bytes
	}

	return database.PutDevices(a.ResinUUID, buffer)
}

func (a *Application) GetDevices() error {
	a.Devices = make(map[string]*device.Device)

	buffer, err := database.GetDevices(a.ResinUUID)
	if err != nil {
		return err
	}

	for _, bytes := range buffer {
		d, err := device.Unmarshall(bytes)
		if err != nil {
			return err
		}

		a.Devices[d.LocalUUID] = d
	}

	return nil
}

func init() {
	log.SetLevel(config.GetLogLevel())

	List = make(map[int]*Application)

	log.Debug("Initialised applications")
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

func (a *Application) handleDeleteFlag() error {
	for key, d := range a.Devices {
		if !d.DeleteFlag {
			continue
		}

		delete(a.Devices, key)

		log.WithFields(log.Fields{
			"name": d.Name,
		}).Info("Device deleted")
	}

	return nil
}

func (a *Application) handleRestartFlag() error {
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
			"name": d.Name,
		}).Info("Device restarted")
	}

	return nil
}
