package database

import (
	"bytes"
	"encoding/binary"
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

// PutDevice puts a specific device
func PutDevice(applicationUUID int, localUUID, deviceUUID string, device []byte) error {
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

	return putDeviceMapping(db, applicationUUID, localUUID, deviceUUID)
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

// GetDeviceMapping gets the applicationUUID and localUUID for a specific device
func GetDeviceMapping(deviceUUID string) (int, string, error) {
	db, err := open()
	if err != nil {
		return 0, "", err
	}
	defer db.Close()

	var applicationUUID []byte
	var localUUID []byte
	if err = db.View(func(tx *bolt.Tx) error {
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

		value = d.Get([]byte("localUUID"))
		if value == nil {
			return fmt.Errorf("Value not found")
		}
		localUUID = make([]byte, len(value))
		copy(localUUID, value)

		return nil
	}); err != nil {
		return 0, "", err
	}

	a, err := b2i(applicationUUID)
	if err != nil {
		return 0, "", err
	}

	return a, (string)(localUUID), nil
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

func putDeviceMapping(db *bolt.DB, applicationUUID int, localUUID, deviceUUID string) error {
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

		if err := d.Put([]byte("applicationUUID"), converted); err != nil {
			return err
		}

		return d.Put([]byte("localUUID"), []byte(localUUID))
	})
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
