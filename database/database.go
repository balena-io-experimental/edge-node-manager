package database

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/resin-io/edge-node-manager/config"
)

var dbPath string

func PutDevice(applicationUUID int, localUUID, deviceUUID string, device []byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	if err = db.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		if b = tx.Bucket([]byte("Applications")); b == nil {
			return fmt.Errorf("Bucket not found")
		}

		var converted []byte
		if converted, err = i2b(applicationUUID); err != nil {
			return err
		}

		var a *bolt.Bucket
		if a, err = b.CreateBucketIfNotExists(converted); err != nil {
			return err
		}

		return a.Put([]byte(deviceUUID), device)
	}); err != nil {
		return err
	}

	return putDeviceMapping(db, applicationUUID, localUUID, deviceUUID)
}

func PutDevices(applicationUUID int, devices map[string][]byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		if b = tx.Bucket([]byte("Applications")); b == nil {
			return fmt.Errorf("Bucket not found")
		}

		var converted []byte
		if converted, err = i2b(applicationUUID); err != nil {
			return err
		}

		var a *bolt.Bucket
		if a, err = b.CreateBucketIfNotExists(converted); err != nil {
			return err
		}

		for key, value := range devices {
			if err = a.Put([]byte(key), value); err != nil {
				return err
			}
		}

		return nil
	})
}

func GetDevice(applicationUUID int, deviceUUID string) ([]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var device []byte
	if err = db.View(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		if b = tx.Bucket([]byte("Applications")); b == nil {
			return fmt.Errorf("Bucket not found")
		}

		var converted []byte
		if converted, err = i2b(applicationUUID); err != nil {
			return err
		}

		var a *bolt.Bucket
		if a = b.Bucket(converted); a == nil {
			return fmt.Errorf("Bucket not found")
		}

		var value []byte
		if value = a.Get([]byte(deviceUUID)); value == nil {
			return fmt.Errorf("Value not found")
		}

		device = make([]byte, len(value))
		copy(device, value)

		return nil
	}); err != nil {
		return nil, err
	}

	return device, nil
}

func GetDevices(applicationUUID int) (map[string][]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var devices map[string][]byte
	if err = db.View(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		if b = tx.Bucket([]byte("Applications")); b == nil {
			return fmt.Errorf("Bucket not found")
		}

		var converted []byte
		if converted, err = i2b(applicationUUID); err != nil {
			return err
		}

		var a *bolt.Bucket
		if a = b.Bucket(converted); a == nil {
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
	}); err != nil {
		return nil, err
	}

	return devices, nil
}

func DeleteDevice(applicationUUID int, deviceUUID string) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	fmt.Println(applicationUUID)
	fmt.Println(deviceUUID)

	if err := db.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		if b = tx.Bucket([]byte("Applications")); b == nil {
			return fmt.Errorf("Bucket not found")
		}
		fmt.Println("here1")

		var converted []byte
		if converted, err = i2b(applicationUUID); err != nil {
			return err
		}

		fmt.Println("here2")

		var a *bolt.Bucket
		if a = b.Bucket(converted); a == nil {
			return nil
		}

		fmt.Println("here3")

		return a.Delete([]byte(deviceUUID))
	}); err != nil {
		return err
	}

	fmt.Println("here4")

	return deleteDeviceMapping(db, deviceUUID)
}

func GetDeviceMapping(deviceUUID string) (int, string, error) {
	db, err := open()
	if err != nil {
		return 0, "", err
	}
	defer db.Close()

	var applicationUUID []byte
	var localUUID []byte
	if err = db.View(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		if b = tx.Bucket([]byte("Mapping")); b == nil {
			return fmt.Errorf("Bucket not found")
		}

		var d *bolt.Bucket
		if d = b.Bucket([]byte(deviceUUID)); d == nil {
			return fmt.Errorf("Bucket not found")
		}

		var value []byte
		if value = d.Get([]byte("applicationUUID")); value == nil {
			return fmt.Errorf("Value not found")
		}

		applicationUUID = make([]byte, len(value))
		copy(applicationUUID, value)

		if value = d.Get([]byte("localUUID")); value == nil {
			return fmt.Errorf("Value not found")
		}

		localUUID = make([]byte, len(value))
		copy(localUUID, value)

		return nil
	}); err != nil {
		return 0, "", err
	}

	var a int
	if a, err = b2i(applicationUUID); err != nil {
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
		var b *bolt.Bucket
		if b = tx.Bucket([]byte("Mapping")); b == nil {
			return fmt.Errorf("Bucket not found")
		}

		var d *bolt.Bucket
		var err error
		if d, err = b.CreateBucketIfNotExists([]byte(deviceUUID)); err != nil {
			return err
		}

		var converted []byte
		if converted, err = i2b(applicationUUID); err != nil {
			return err
		}

		if err = d.Put([]byte("applicationUUID"), converted); err != nil {
			return err
		}

		return d.Put([]byte("localUUID"), []byte(localUUID))
	})
}

func deleteDeviceMapping(db *bolt.DB, deviceUUID string) error {
	return db.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		if b = tx.Bucket([]byte("Mapping")); b == nil {
			return fmt.Errorf("Bucket not found")
		}

		return b.Delete([]byte(deviceUUID))
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
