package database

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"path"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/config"

	"github.com/boltdb/bolt"
)

var dbPath string

func PutDevice(applicationUUID int, deviceUUID string, device []byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Applications"))

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
}

func GetDevice(applicationUUID int, deviceUUID string) ([]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var device []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Applications"))

		converted, err := i2b(applicationUUID)
		if err != nil {
			return err
		}

		a := b.Bucket(converted)
		if a == nil {
			return nil
		}

		buffer := a.Get([]byte(deviceUUID))
		device := make([]byte, len(buffer))
		copy(device, buffer)

		return nil
	})

	return device, nil
}

func GetDevices(applicationUUID int) (map[string][]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var devices map[string][]byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Applications"))

		converted, err := i2b(applicationUUID)
		if err != nil {
			return err
		}

		a := b.Bucket(converted)
		if a == nil {
			return nil
		}

		devices := make(map[string][]byte)
		a.ForEach(func(k, v []byte) error {
			key := make([]byte, len(k))
			value := make([]byte, len(v))
			copy(key, k)
			copy(value, v)
			devices[(string)(key)] = value
			return nil
		})

		return nil
	})

	return devices, nil
}

func PutDeviceField(applicationUUID int, deviceUUID, field string, value []byte) error {
	buffer, err := unmarshall(applicationUUID, deviceUUID)
	if err != nil {
		return err
	}

	buffer[field] = value

	return marshall(applicationUUID, deviceUUID, buffer)
}

func GetDeviceField(applicationUUID int, deviceUUID, field string) ([]byte, error) {
	buffer, err := unmarshall(applicationUUID, deviceUUID)
	if err != nil {
		return nil, err
	}

	return buffer[field].([]byte), nil
}

func PutDeviceMapping(applicationUUID int, deviceUUID string) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Mapping"))

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

func GetDeviceMapping(deviceUUID string) (int, error) {
	db, err := open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var applicationUUID []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Mapping"))

		d := b.Bucket([]byte(deviceUUID))
		if d == nil {
			return nil
		}

		buffer := d.Get([]byte("applicationUUID"))
		applicationUUID := make([]byte, len(buffer))
		copy(applicationUUID, buffer)

		return nil
	})

	converted, err := b2i(applicationUUID)
	if err != nil {
		return 0, err
	}

	return converted, nil
}

func init() {
	dir := config.GetDbDir()
	name := config.GetDbName()
	dbPath = path.Join(dir, name)

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

func marshall(applicationUUID int, deviceUUID string, buffer map[string]interface{}) error {
	bytes, err := json.Marshal(buffer)
	if err != nil {
		return err
	}

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
