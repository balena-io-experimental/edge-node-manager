package database

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/structs"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/device"
)

/*
Uses the tiedot database
https://github.com/HouzuoGuo/tiedot
*/

var directory string
var connection *db.DB

func LoadDevices(uuid string, t *device.Type) (map[int]device.Interface, error) {
	initialise()
	query := fmt.Sprintf(`[{"eq": "%s", "in": ["ApplicationUUID"]}]`, uuid)
	var q interface{}
	if err := json.Unmarshal([]byte(query), &q); err != nil {
		return nil, err
	}

	queryResult := make(map[int]struct{})
	devices := connection.Use("Devices")
	if err := db.EvalQuery(q, devices, &queryResult); err != nil {
		return nil, err
	}

	result := make(map[int]device.Interface)
	for key := range queryResult {
		if device, err := loadDevice(key, t); err != nil {
			return nil, err
		} else {
			result[key] = device
		}
	}

	return result, nil
}

func SaveDevice(d device.Interface) (int, error) {
	devices := connection.Use("Devices")
	return devices.Insert(interfaceToMap(d))
}

func RemoveDevice(key int) error {
	devices := connection.Use("Devices")
	return devices.Delete(key)
}

func UpdateDevices(d map[int]device.Interface) error {
	initialise()
	for key, value := range d {
		if err := updateDevice(key, value); err != nil {
			return err
		}
	}

	return nil
}

func Stop() {
	if connection != nil {
		if err := connection.Close(); err != nil {
			log.Fatalf("Unable to close database connection: %v", err)
		}
	}
}

func initialise() {
	if directory == "" {
		directory = config.GetDbDirectory()
	}
	if connection == nil {
		var err error
		if connection, err = db.OpenDB(directory); err != nil {
			log.Fatalf("Unable to open database connection: %v", err)
		} else {
			collection := createCollection("Devices")
			createIndex("ApplicationUUID", collection)
		}
		log.Printf("Opened database connection to %s", directory)
	}
}

func createCollection(name string) *db.Col {
	exists := false
	collections := connection.AllCols()
	for _, collection := range collections {
		if collection == name {
			exists = true
			break
		}
	}
	if !exists {
		connection.Create(name)
	}
	return connection.Use(name)
}

func createIndex(name string, collection *db.Col) {
	exists := false
	paths := collection.AllIndexes()
	for _, path := range paths {
		for _, index := range path {
			if index == name {
				exists = true
				break
			}
		}
		if exists {
			break
		}
	}
	if !exists {
		collection.Index([]string{name})
	}
}

func loadDevice(key int, t *device.Type) (device.Interface, error) {
	devices := connection.Use("Devices")
	if readBack, err := devices.Read(key); err != nil {
		return nil, err
	} else {
		if bytes, err := json.Marshal(readBack); err != nil { // How can we avoid marshalling and then unmarshalling to populate the device
			return nil, err
		} else {
			device := device.Create(t)
			json.Unmarshal(bytes, device.GetDevice()) // See comment above
			return device, nil
		}
	}
}

func updateDevice(key int, d device.Interface) error {
	devices := connection.Use("Devices")
	return devices.Update(key, interfaceToMap(d))
}

func interfaceToMap(d device.Interface) map[string]interface{} {
	device := structs.Map(d.GetDevice())
	delete(device, "Type")
	return device
}
