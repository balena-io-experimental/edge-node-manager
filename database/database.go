package database

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/josephroberts/edge-node-manager/config"
)

// Uses the bolt package
// https://github.com/boltdb/bolt

var dbPath string

// There are two buckets in use "Applications" and "Mapping"
// "Applications" contains a bucket of applications, where the applicationUUID is the key
// Each application bucket contains a bucket of devices, where the deviceUUID is the key
// "Mapping" contains the mapping between deviceUUID and applicationUUID
// Where the deviceUUID is the key and applicationUUID is the value

// PutDevice puts a specific device
func PutDevice(applicationUUID int, deviceUUID string, device []byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Applications"))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		converted, err := i2b(applicationUUID)
		if err != nil {
			return err
		}

		a, err := b.CreateBucketIfNotExists(converted)
		if err != nil {
			return err
		}

		return a.Put([]byte(deviceUUID), device)
	})
	if err != nil {
		return err
	}

	return putDeviceMapping(db, applicationUUID, deviceUUID)
}

// PutDevices puts all devices associated to a specific application
func PutDevices(applicationUUID int, devices map[string][]byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Applications"))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		converted, err := i2b(applicationUUID)
		if err != nil {
			return err
		}

		a, err := b.CreateBucketIfNotExists(converted)
		if err != nil {
			return err
		}

		for key, value := range devices {
			if err = a.Put([]byte(key), value); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// GetDevice gets a specific device
func GetDevice(applicationUUID int, deviceUUID string) ([]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var device []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Applications"))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		converted, err := i2b(applicationUUID)
		if err != nil {
			return err
		}

		a := b.Bucket(converted)
		if a == nil {
			return fmt.Errorf("Bucket not found")
		}

		value := a.Get([]byte(deviceUUID))
		if value == nil {
			return fmt.Errorf("Value not found")
		}

		device = make([]byte, len(value))
		copy(device, value)

		return nil
	})

	return device, err
}

// GetDevices gets all devices associated to a specific application
func GetDevices(applicationUUID int) (map[string][]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var devices map[string][]byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Applications"))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		converted, err := i2b(applicationUUID)
		if err != nil {
			return err
		}

		a := b.Bucket(converted)
		if a == nil {
			return nil
		}

		devices = make(map[string][]byte)
		return a.ForEach(func(k, v []byte) error {
			key := make([]byte, len(k))
			value := make([]byte, len(v))
			copy(key, k)
			copy(value, v)
			devices[(string)(key)] = value
			return nil
		})
	})

	return devices, err
}

// PutDeviceField puts a field for a specific device
func PutDeviceField(applicationUUID int, deviceUUID, field string, value []byte) error {
	buffer, err := unmarshall(applicationUUID, deviceUUID)
	if err != nil {
		return err
	}

	buffer[field] = value

	log.WithFields(log.Fields{
		"field": field,
		"value": value,
		"value": (string)(value),
	}).Debug("Put device field")

	return marshall(applicationUUID, deviceUUID, buffer)
}

// GetDeviceField gets a field for a specific device
func GetDeviceField(applicationUUID int, deviceUUID, field string) ([]byte, error) {
	buffer, err := unmarshall(applicationUUID, deviceUUID)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"bytes": buffer,
	}).Debug("Unmarshall")

	if v, ok := buffer[field].(string); ok {
		return ([]byte)(v), nil
	} else if v, ok := buffer[field].(int); ok {
		return i2b(v)
	}

	return nil, fmt.Errorf("Type not supported")
}

// GetDeviceMapping gets the applicationUUID for a specific device
func GetDeviceMapping(deviceUUID string) (int, error) {
	db, err := open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var applicationUUID []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Mapping"))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		d := b.Bucket([]byte(deviceUUID))
		if d == nil {
			return fmt.Errorf("Bucket not found")
		}

		value := d.Get([]byte("applicationUUID"))
		if value == nil {
			return fmt.Errorf("Value not found")
		}

		applicationUUID = make([]byte, len(value))
		copy(applicationUUID, value)

		return nil
	})
	if err != nil {
		return 0, err
	}

	return b2i(applicationUUID)
}

func init() {
	log.SetLevel(config.GetLogLevel())

	dir := config.GetDbDir()
	name := config.GetDbName()
	dbPath = path.Join(dir, name)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"Path":  dir,
			"Error": err,
		}).Fatal("Unable to create path")
	}

	db, err := open()
	if err != nil {
		log.WithFields(log.Fields{
			"Path":  dbPath,
			"Error": err,
		}).Fatal("Unable to open database connection")
	}
	defer db.Close()

	if err := makeBucket(db, "Applications"); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Create bucket failed")
	}

	if err := makeBucket(db, "Mapping"); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Create bucket failed")
	}

	log.WithFields(log.Fields{
		"Path": dbPath,
	}).Debug("Opened database connection")
}

func open() (*bolt.DB, error) {
	return bolt.Open(dbPath, 0600, nil)
}

func makeBucket(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
}

func putDeviceMapping(db *bolt.DB, applicationUUID int, deviceUUID string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Mapping"))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		d, err := b.CreateBucketIfNotExists([]byte(deviceUUID))
		if err != nil {
			return err
		}

		converted, err := i2b(applicationUUID)
		if err != nil {
			return err
		}

		return d.Put([]byte("applicationUUID"), converted)
	})
}

func marshall(applicationUUID int, deviceUUID string, buffer map[string]interface{}) error {
	bytes, err := json.Marshal(buffer)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"bytes": (string)(bytes),
	}).Debug("Marshall")

	return PutDevice(applicationUUID, deviceUUID, bytes)
}

func unmarshall(applicationUUID int, deviceUUID string) (map[string]interface{}, error) {
	bytes, err := GetDevice(applicationUUID, deviceUUID)
	if err != nil {
		return nil, err
	}

	buffer := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}

func i2b(value int) ([]byte, error) {
	result := new(bytes.Buffer)

	if err := binary.Write(result, binary.LittleEndian, (int32)(value)); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func b2i(value []byte) (int, error) {
	var result int32

	if err := binary.Read(bytes.NewReader(value), binary.LittleEndian, &result); err != nil {
		return 0, err
	}

	return (int)(result), nil
}
