package database

import (
	"encoding/json"
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"

	tiedotDb "github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/structs"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/device"
)

// Uses the tiedot database
// https://github.com/HouzuoGuo/tiedot

var (
	directory  string
	connection *tiedotDb.DB
)

// LoadDevices loads and returns all devices from the database for a specific application
func LoadDevices(appUUID int) (map[string]*device.Device, error) { // TODO pointers
	results, err := query(strconv.Itoa(appUUID), "applicationUUID")
	if err != nil {
		return nil, err
	}

	devices := make(map[string]*device.Device)
	for result := range results {
		device, err := loadDevice(result)
		if err != nil {
			return nil, err
		}
		devices[device.LocalUUID] = device
	}

	return devices, nil
}

// SaveDevice saves a new device to the database and returns the new device
func SaveDevice(newDevice *device.Device) (*device.Device, error) {
	collection := connection.Use("Devices")
	key, err := collection.Insert(structs.Map(newDevice))
	if err != nil {
		return &device.Device{}, err
	}

	return loadDevice(key)
}

// RemoveDevice deletes a device from the database
func RemoveDevice(key int) error {
	collection := connection.Use("Devices")
	return collection.Delete(key)
}

// UpdateDevices updates devices in the database
func UpdateDevices(existingDevices map[string]*device.Device) error {
	for _, existingDevice := range existingDevices {
		if err := updateDevice(existingDevice); err != nil {
			return err
		}
	}

	return nil
}

// SetTargetCommit sets the target commit for a specific device
func SetTargetCommit(ResinUUID, commit string) error {
	results, err := query(ResinUUID, "resinUUID")
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return fmt.Errorf("Device not found in the database")
	} else if len(results) > 1 {
		return fmt.Errorf("More than one device found in the database")
	}

	collection := connection.Use("Devices")
	for result := range results {
		if err := collection.Update(result, map[string]interface{}{"targetCommit": commit}); err != nil {
			return err
		}
	}

	return nil
}

// Stop closes the database connection
func Stop() error {
	if connection == nil {
		return nil
	}
	return connection.Close()
}

func init() {
	directory = config.GetDbDir()

	var err error
	if connection, err = tiedotDb.OpenDB(directory); err != nil {
		log.WithFields(log.Fields{
			"Directory": directory,
			"Error":     err,
		}).Fatal("Unable to open database connection")
	} else {
		// Ignore error as the Devices collection could already exist
		connection.Create("Devices")
		collection := connection.Use("Devices")
		// Ignore error as the applicationUUID index could already exist
		collection.Index([]string{"applicationUUID"})
		// Ignore error as the resinUUID index could already exist
		collection.Index([]string{"resinUUID"})
	}

	log.WithFields(log.Fields{
		"Directory": directory,
	}).Debug("Opened database connection")
}

func query(value, field string) (map[int]struct{}, error) {
	query := fmt.Sprintf(`[{"eq": "%s", "in": ["%s"]}]`, value, field)

	// TODO: Research how to avoid this unmarshalling step
	var q interface{}
	if err := json.Unmarshal([]byte(query), &q); err != nil {
		return nil, err
	}

	queryResult := make(map[int]struct{})
	collection := connection.Use("Devices")
	if err := tiedotDb.EvalQuery(q, collection, &queryResult); err != nil {
		return nil, err
	}

	return queryResult, nil
}

func loadDevice(key int) (*device.Device, error) {
	collection := connection.Use("Devices")
	readBack, err := collection.Read(key)
	if err != nil {
		return &device.Device{}, err
	}

	// Set the DatabaseUUID
	// This is necessary as the DB does not store the DatabaseUUID field correctly
	// Save 4170124961882522202, and get 1.1229774266282973e+18 back (looks like overflow)
	readBack["databaseUUID"] = key

	// TODO: Research how to avoid this marshalling step
	// there should ne a nice way to convert a map to a struct, try https://github.com/mitchellh/mapstructure
	bytes, err := json.Marshal(readBack)
	if err != nil {
		return &device.Device{}, err
	}

	existingDevice := &device.Device{}
	if err := json.Unmarshal(bytes, existingDevice); err != nil {
		return &device.Device{}, err
	}

	return existingDevice, nil
}

func updateDevice(existingDevice *device.Device) error {
	collection := connection.Use("Devices")
	return collection.Update(existingDevice.DatabaseUUID, structs.Map(existingDevice))
}
