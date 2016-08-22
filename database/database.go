package database

import (
	"encoding/json"
	"fmt"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/structs"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/radio"
)

/*
Uses the tiedot database
https://github.com/HouzuoGuo/tiedot
*/

type Interface interface {
	Start() error
	Stop() error
	Create(i device.Interface) (int, error)
	Remove(key int) error
	Get(key int, deviceType device.SupportedDevice, radio radio.Interface) (device.Interface, error)
	Update(key int, i device.Interface) error
	Query(field, value string, deviceType device.SupportedDevice, radio radio.Interface) (map[int]device.Interface, error)
}

type Database struct {
	Directory  string
	connection *db.DB
}

func (d *Database) Start() error {
	if temp, err := db.OpenDB(d.Directory); err != nil {
		return err
	} else {
		/*
			For some reason I cannot assign directly to d.Connection
			http://stackoverflow.com/questions/21345274/go-fails-to-infer-type-in-assignment-non-name-on-left-side-of
		*/
		d.connection = temp
	}

	return d.initialise()
}

func (d *Database) Stop() error {
	err := d.connection.Close()
	return err
}

func (d *Database) initialise() error {
	if err := d.connection.Create("Devices"); err != nil {
		return err
	}

	devices := d.connection.Use("Devices")
	if err := devices.Index([]string{"ApplicationUUID"}); err != nil {
		return err
	}
	return devices.Index([]string{"LocalUUID"})
}

func (d *Database) Create(i device.Interface) (int, error) {
	devices := d.connection.Use("Devices")
	return devices.Insert(structs.Map(i.GetDevice()))
}

func (d *Database) Remove(key int) error {
	devices := d.connection.Use("Devices")
	return devices.Delete(key)
}

func (d *Database) Get(key int, deviceType device.SupportedDevice, radio radio.Interface) (device.Interface, error) {
	devices := d.connection.Use("Devices")
	if readBack, err := devices.Read(key); err != nil {
		return nil, err
	} else {
		if b, err := json.Marshal(readBack); err != nil {
			return nil, err
		} else {
			i := device.Create(deviceType)
			i.GetDevice().Deserialise(b)
			i.GetDevice().Radio = radio
			return i, nil
		}
	}
}

func (d *Database) Update(key int, i device.Interface) error {
	devices := d.connection.Use("Devices")
	return devices.Update(key, structs.Map(i.GetDevice()))
}

func (d *Database) Query(field, value string, deviceType device.SupportedDevice, radio radio.Interface) (map[int]device.Interface, error) {
	query := fmt.Sprintf(`[{"eq": "%s", "in": ["%s"]}]`, value, field)
	var q interface{}
	json.Unmarshal([]byte(query), &q)

	queryResult := make(map[int]struct{})
	devices := d.connection.Use("Devices")
	if err := db.EvalQuery(q, devices, &queryResult); err != nil {
		return nil, err
	}

	result := make(map[int]device.Interface)
	for key := range queryResult {
		if i, err := d.Get(key, deviceType, radio); err != nil {
			return nil, err
		} else {
			result[key] = i
		}
	}

	return result, nil

}
