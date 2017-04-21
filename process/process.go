package process

import (
	"os"
	"path"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/fredli74/lockfile"
	"github.com/resin-io/edge-node-manager/application"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/device"
	deviceStatus "github.com/resin-io/edge-node-manager/device/status"
	processStatus "github.com/resin-io/edge-node-manager/process/status"
	"github.com/resin-io/edge-node-manager/supervisor"
	tarinator "github.com/verybluebot/tarinator-go"
)

var (
	CurrentStatus processStatus.Status
	TargetStatus  processStatus.Status
	updateRetries int
	pauseDelay    time.Duration
	lockLocation  string
	lock          *lockfile.LockFile
)

func Run(a application.Application) []error {
	log.Info("----------------------------------------")

	// Pause the process if necessary
	if err := pause(); err != nil {
		return []error{err}
	}

	// Initialise the radio
	if err := a.Board.InitialiseRadio(); err != nil {
		return []error{err}
	}
	defer a.Board.CleanupRadio()

	if log.GetLevel() == log.DebugLevel {
		log.WithFields(log.Fields{
			"Application": a,
		}).Debug("Processing application")
	} else {
		log.WithFields(log.Fields{
			"Application": a.Name,
		}).Info("Processing application")
	}

	// Enable update locking
	var err error
	lock, err = lockfile.Lock(lockLocation)
	if err != nil {
		return []error{err}
	}
	defer lock.Unlock()

	// Handle delete flags
	if err := handleDelete(a); err != nil {
		return []error{err}
	}

	// Get all online devices associated with this application
	onlineDevices, err := getOnlineDevices(a)
	if err != nil {
		return []error{err}
	}

	// Get all provisioned devices associated with this application
	provisionedDevices, err := getProvisionedDevices(a)
	if err != nil {
		return []error{err}
	}

	if log.GetLevel() == log.DebugLevel {
		log.WithFields(log.Fields{
			"Provisioned devices": provisionedDevices,
		}).Debug("Processing application")
	} else {
		log.WithFields(log.Fields{
			"Number of provisioned devices": len(provisionedDevices),
		}).Info("Processing application")
	}

	// Convert provisioned devices to a hash map
	hashmap := make(map[string]struct{})
	var s struct{}
	for _, value := range provisionedDevices {
		hashmap[value.LocalUUID] = s
	}

	// Provision all unprovisioned devices associated with this application
	for key := range onlineDevices {
		if _, ok := hashmap[key]; ok {
			// Device already provisioned
			continue
		}

		// Device not already provisioned
		if errs := provisionDevice(a, key); errs != nil {
			return errs
		}
	}

	// Refesh all provisioned devices associated with this application
	provisionedDevices, err = getProvisionedDevices(a)
	if err != nil {
		return []error{err}
	}

	// Sync all provisioned devices associated with this application
	for _, value := range provisionedDevices {
		if errs := value.Sync(); errs != nil {
			return errs
		}

		if err := updateDevice(value); err != nil {
			return []error{err}
		}
	}

	// Refesh all provisioned devices associated with this application
	provisionedDevices, err = getProvisionedDevices(a)
	if err != nil {
		return []error{err}
	}

	// Set state for all provisioned devices associated with this application
	for _, value := range provisionedDevices {
		if _, ok := onlineDevices[value.LocalUUID]; ok {
			value.Status = deviceStatus.IDLE
		} else {
			value.Status = deviceStatus.OFFLINE
		}

		if err := updateDevice(value); err != nil {
			return []error{err}
		}

		if errs := sendState(value); errs != nil {
			return errs
		}
	}

	// Refesh all provisioned devices associated with this application
	provisionedDevices, err = getProvisionedDevices(a)
	if err != nil {
		return []error{err}
	}

	// Update all online, outdated, provisioned devices associated with this application
	for _, value := range provisionedDevices {
		if (value.Commit != value.TargetCommit) && (value.Status != deviceStatus.OFFLINE) {
			// Populate board (and micro) for the device
			if err := value.PopulateBoard(); err != nil {
				return []error{err}
			}

			// Perform the update
			if errs := updateFirmware(value); errs != nil {
				return errs
			}
		}
	}

	return nil
}

func init() {
	log.SetLevel(config.GetLogLevel())

	var err error
	if pauseDelay, err = config.GetPauseDelay(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load pause delay")
	}

	if updateRetries, err = config.GetUpdateRetries(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to update retries")
	}

	lockLocation = config.GetLockFileLocation()

	CurrentStatus = processStatus.RUNNING
	TargetStatus = processStatus.RUNNING

	log.Debug("Initialised process")
}

func pause() error {
	if TargetStatus != processStatus.PAUSED {
		return nil
	}

	CurrentStatus = processStatus.PAUSED
	log.WithFields(log.Fields{
		"Status": CurrentStatus,
	}).Info("Process status")

	for TargetStatus == processStatus.PAUSED {
		time.Sleep(pauseDelay)
	}

	CurrentStatus = processStatus.RUNNING
	log.WithFields(log.Fields{
		"Status": CurrentStatus,
	}).Info("Process status")

	return nil
}

func getOnlineDevices(a application.Application) (map[string]struct{}, error) {
	onlineDevices, err := a.Board.Scan(a.ResinUUID)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"Number of online devices": len(onlineDevices),
	}).Info("Processing application")

	return onlineDevices, nil
}

func getProvisionedDevices(a application.Application) ([]device.Device, error) {
	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var provisionedDevices []device.Device
	if err := db.Find("ApplicationUUID", a.ResinUUID, &provisionedDevices); err != nil && err.Error() != index.ErrNotFound.Error() {
		return nil, err
	}

	return provisionedDevices, nil
}

func provisionDevice(a application.Application, localUUID string) []error {
	log.WithFields(log.Fields{
		"Local UUID": localUUID,
	}).Info("Provisioning device")

	resinUUID, name, errs := supervisor.DependentDeviceProvision(a.ResinUUID)
	if errs != nil {
		return errs
	}

	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		return []error{err}
	}
	defer db.Close()

	d := device.New(a.ResinUUID, a.BoardType, name, localUUID, resinUUID)
	if err := db.Save(&d); err != nil {
		return []error{err}
	}

	log.WithFields(log.Fields{
		"Name":       d.Name,
		"Local UUID": d.LocalUUID,
	}).Info("Provisioned device")

	return nil
}

func updateDevice(d device.Device) error {
	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(&d)
}

func sendState(d device.Device) []error {
	online := true
	if d.Status == deviceStatus.OFFLINE {
		online = false
	}

	return supervisor.DependentDeviceInfoUpdateWithOnlineState(d.ResinUUID, (string)(d.Status), d.Commit, online)
}

func updateFirmware(d device.Device) []error {
	online, err := d.Board.Online()
	if err != nil {
		return []error{err}
	} else if !online {
		return nil
	}

	filepath, err := getFirmware(d)
	if err != nil {
		return []error{err}
	}

	d.Status = deviceStatus.INSTALLING
	if err := updateDevice(d); err != nil {
		return []error{err}
	}
	if errs := sendState(d); errs != nil {
		return errs
	}

	for i := 1; i <= updateRetries; i++ {
		log.WithFields(log.Fields{
			"Name":    d.Name,
			"Attempt": i,
		}).Info("Starting update")

		if err := d.Board.Update(filepath); err != nil {
			log.WithFields(log.Fields{
				"Name":  d.Name,
				"Error": err,
			}).Error("Update failed")
			continue
		} else {
			log.WithFields(log.Fields{
				"Name": d.Name,
			}).Info("Finished update")
			d.Commit = d.TargetCommit
			break
		}
	}

	d.Status = deviceStatus.IDLE
	if err := updateDevice(d); err != nil {
		return []error{err}
	}
	return sendState(d)
}

func getFirmware(d device.Device) (string, error) {
	// Build the file paths
	filepath := config.GetAssetsDir()
	filepath = path.Join(filepath, strconv.Itoa(d.ApplicationUUID))
	filepath = path.Join(filepath, d.TargetCommit)
	tarPath := path.Join(filepath, "binary.tar")

	// Check if the firmware exists
	if _, err := os.Stat(tarPath); os.IsNotExist(err) {
		// Download the firmware
		if err := supervisor.DependentApplicationUpdate(d.ApplicationUUID, d.TargetCommit); err != nil {
			return "", err
		}

		// Extract the firmware
		if err := tarinator.UnTarinate(filepath, tarPath); err != nil {
			return "", err
		}
	}

	return filepath, nil
}

func handleDelete(a application.Application) error {
	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Select(q.Eq("ApplicationUUID", a.ResinUUID), q.Eq("DeleteFlag", true)).Delete(&device.Device{}); err != nil && err.Error() != index.ErrNotFound.Error() {
		return err
	}

	return nil
}
