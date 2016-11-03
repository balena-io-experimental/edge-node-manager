package database

import (
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

var dbPath string

// Write a map to the database
// p is the outer bucket key (applicationUUID)
// c is the inner bucket key(resinUUID)
// m is the key/value map
func Write(p, c string, m map[string][]byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("applications"))
		if err != nil {
			return err
		}

		parent, err := root.CreateBucketIfNotExists([]byte(p))
		if err != nil {
			return err
		}

		child, err := parent.CreateBucketIfNotExists([]byte(c))
		if err != nil {
			return err
		}

		for k, v := range m {
			if err := child.Put([]byte(k), v); err != nil {
				return err
			}
		}

		return nil
	})
}

// Writes an array of maps to the database
// p is the outer bucket key (applicationUUID)
// cf is the inner bucket key field (resinUUID)
// m is the array of key/value maps
func Writes(p, cf string, m []map[string][]byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("applications"))
		if err != nil {
			return err
		}

		parent, err := root.CreateBucketIfNotExists([]byte(p))
		if err != nil {
			return err
		}

		for _, v := range m {
			child, err := parent.CreateBucketIfNotExists(v[cf])
			if err != nil {
				return err
			}

			for ik, iv := range v {
				if err := child.Put([]byte(ik), iv); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// WriteField writes a field to the database
// p is the outer bucket key (applicationUUID)
// c is the inner bucket key(resinUUID)
// k is the field key
// v is the field value
func WriteField(p, c, k string, v []byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("applications"))
		if err != nil {
			return err
		}

		parent, err := root.CreateBucketIfNotExists([]byte(p))
		if err != nil {
			return err
		}

		child, err := parent.CreateBucketIfNotExists([]byte(c))
		if err != nil {
			return err
		}

		return child.Put([]byte(k), v)
	})
}

// WritesField writes a field to the child maps in the database
// p is the outer bucket key (applicationUUID)
// k is the field key
// v is the field value
func WritesField(p, k string, v []byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("applications"))
		if err != nil {
			return err
		}

		parent, err := root.CreateBucketIfNotExists([]byte(p))
		if err != nil {
			return err
		}

		i := parent.Cursor()
		for ik, _ := i.First(); ik != nil; ik, _ = i.Next() {
			child := parent.Bucket(ik)

			if err := child.Put([]byte(k), v); err != nil {
				return err
			}
		}

		return nil
	})
}

// WriteMapping writes key/value pairs associated with p
// p is the outer bucket key (resinUUID)
func WriteMapping(p string, m map[string][]byte) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("mapping"))
		if err != nil {
			return err
		}

		parent, err := root.CreateBucketIfNotExists([]byte(p))
		if err != nil {
			return err
		}

		for k, v := range m {
			err = parent.Put([]byte(k), v)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Read a map from the database
// p is the outer bucket key (applicationUUID)
// c is the inner bucket key (resinUUID)
// Returns map key/value
func Read(p, c string) (map[string][]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	m := make(map[string][]byte)

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("applications"))
		if root == nil {
			return err
		}

		parent := root.Bucket([]byte(p))
		if parent == nil {
			return fmt.Errorf("Bucket %s not found", (string)(p))
		}

		child := parent.Bucket([]byte(c))
		if child == nil {
			return fmt.Errorf("Bucket %s not found", (string)(c))
		}

		i := child.Cursor()
		for k, v := i.First(); k != nil; k, v = i.Next() {
			value := make([]byte, len(v))
			copy(value, v)
			m[(string)(k)] = value
		}

		return nil
	})

	return m, err
}

// Reads an array of maps from the database
// p is the outer bucket key (applicationUUID)
// Returns array of maps key/value
func Reads(p string) ([]map[string][]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var a []map[string][]byte

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("applications"))
		if root == nil {
			return err
		}

		parent := root.Bucket([]byte(p))
		if parent == nil {
			return fmt.Errorf("Bucket %s not found", (string)(p))
		}

		i := parent.Cursor()
		for k, _ := i.First(); k != nil; k, _ = i.Next() {
			child := parent.Bucket(k)
			if child == nil {
				return fmt.Errorf("Bucket %s not found", (string)(k))
			}

			m := make(map[string][]byte)

			ii := child.Cursor()
			for ik, iv := ii.First(); ik != nil; ik, iv = ii.Next() {
				value := make([]byte, len(iv))
				copy(value, iv)
				m[(string)(ik)] = value
			}

			a = append(a, m)
		}

		return nil
	})

	return a, err
}

// ReadField reads a field from the database
// p is the outer bucket key (applicationUUID)
// c is the inner bucket key (resinUUID)
// k is the field key
// Returns field value
func ReadField(p, c, k string) ([]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var a []byte

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("applications"))
		if root == nil {
			return err
		}

		parent := root.Bucket([]byte(p))
		if parent == nil {
			return fmt.Errorf("Bucket %s not found", (string)(p))
		}

		child := parent.Bucket([]byte(c))
		if child == nil {
			return fmt.Errorf("Bucket %s not found", (string)(c))
		}

		v := child.Get([]byte(k))

		value := make([]byte, len(v))
		copy(value, v)
		a = value

		return nil
	})

	return a, err
}

// ReadsField reads a field from the child maps in the database
// p is the outer bucket key (applicationUUID)
// k is the field key
// Returns field map key/value
func ReadsField(p, k string) (map[string][]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	m := make(map[string][]byte)

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("applications"))
		if root == nil {
			return err
		}

		parent := root.Bucket([]byte(p))
		if parent == nil {
			return fmt.Errorf("Bucket %s not found", (string)(p))
		}

		i := parent.Cursor()
		for ik, _ := i.First(); ik != nil; ik, _ = i.Next() {
			child := parent.Bucket(ik)
			if child == nil {
				return fmt.Errorf("Bucket %s not found", (string)(ik))
			}

			v := child.Get([]byte(k))

			value := make([]byte, len(v))
			copy(value, v)
			m[(string)(ik)] = value
		}

		return nil
	})

	return m, err
}

// ReadMapping reads key/value pairs associated with p
// p is the outer bucket key (resinUUID)
// Returns field map key/value
func ReadMapping(p string) (map[string][]byte, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	m := make(map[string][]byte)

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("mapping"))
		if root == nil {
			return err
		}

		parent := root.Bucket([]byte(p))
		if parent == nil {
			return fmt.Errorf("Bucket %s not found", (string)(p))
		}

		i := parent.Cursor()
		for k, v := i.First(); k != nil; k, v = i.Next() {
			value := make([]byte, len(v))
			copy(value, v)
			m[(string)(k)] = value
		}

		return nil
	})

	return m, err
}

// Delete a map from the database
// p is the outer bucket key (applicationUUID)
// c is the inner bucket key (resinUUID)
func Delete(p, c string) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("applications"))
		if err != nil {
			return err
		}

		parent, err := root.CreateBucketIfNotExists([]byte(p))
		if err != nil {
			return err
		}

		return parent.DeleteBucket([]byte(c))
	})
}

// Deletes maps from the database
// p is the outer bucket key (applicationUUID)
func Deletes(p string) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("applications"))
		if err != nil {
			return err
		}

		return root.DeleteBucket([]byte(p))
	})
}

// DeleteMapping deletes key/value pairs associated with p
// p is the outer bucket key (resinUUID)
func DeleteMapping(p string) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("mapping"))
		if err != nil {
			return err
		}

		return root.DeleteBucket([]byte(p))
	})
}

func initialise(d, n string) {
	if err := os.MkdirAll(d, os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"Path":  dbPath,
			"Error": err,
		}).Fatal("Unable to create database path")
	}

	dbPath = path.Join(d, n)
}

func open() (*bolt.DB, error) {
	return bolt.Open(dbPath, 0600, nil)
}
