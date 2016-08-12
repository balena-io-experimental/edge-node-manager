package database

import (
	"flag"
	"os"
	"testing"

	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/radio"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestCreate(t *testing.T) {
	dir := "/tmp/MyDatabase"
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	db := &Database{Directory: dir}
	if err := db.Start(); err != nil {
		t.Error("Start db failed")
		t.Error(err)
	}
	defer db.Stop()

	d := device.Create(device.NRF51822)
	r := radio.Create(radio.BLUETOOTH)
	d.GetDevice().ApplicationUUID = "application"
	d.GetDevice().LocalUUID = "local"
	d.GetDevice().Radio = r

	if key, err := db.Create(d); err != nil {
		t.Error("Create device failed")
		t.Error(err)
	} else if i, err := db.Get(key, device.NRF51822, r); err != nil {
		t.Error("Create device failed")
		t.Error(err)
	} else {
		if d.GetDevice().ApplicationUUID != i.GetDevice().ApplicationUUID {
			t.Error("expected" + d.GetDevice().ApplicationUUID)
		}
		if d.GetDevice().LocalUUID != i.GetDevice().LocalUUID {
			t.Error("expected" + d.GetDevice().LocalUUID)
		}
		if d.GetDevice().Radio != i.GetDevice().Radio {
			t.Errorf(`expected %s`, d.GetDevice().Radio)
		}
	}
}

func TestRemove(t *testing.T) {
	dir := "/tmp/MyDatabase"
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	db := &Database{Directory: dir}
	if err := db.Start(); err != nil {
		t.Error("Start db failed")
		t.Error(err)
	}
	defer db.Stop()

	d := device.Create(device.NRF51822)
	r := radio.Create(radio.BLUETOOTH)
	d.GetDevice().ApplicationUUID = "application"
	d.GetDevice().LocalUUID = "local"
	d.GetDevice().Radio = r

	if key, err := db.Create(d); err != nil {
		t.Error("Remove device failed")
		t.Error(err)
	} else if _, err := db.Get(key, device.NRF51822, r); err != nil {
		t.Error("Remove device failed")
		t.Error(err)
	} else if err := db.Remove(key); err != nil {
		t.Error("Remove device failed")
		t.Error(err)
	} else if _, err := db.Get(key, device.NRF51822, r); err == nil {
		t.Error("Remove device failed")
		t.Error(err)
	}
}

func TestUpdate(t *testing.T) {
	dir := "/tmp/MyDatabase"
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	db := &Database{Directory: dir}
	if err := db.Start(); err != nil {
		t.Error("Start db failed")
		t.Error(err)
	}
	defer db.Stop()

	d := device.Create(device.NRF51822)
	r := radio.Create(radio.BLUETOOTH)
	d.GetDevice().ApplicationUUID = "application"
	d.GetDevice().LocalUUID = "local"
	d.GetDevice().Radio = r

	key, err := db.Create(d)
	if err != nil {
		t.Error("Update device failed")
		t.Error(err)
	}
	if i, err := db.Get(key, device.NRF51822, r); err != nil {
		t.Error("Update device failed")
		t.Error(err)
	} else {
		if d.GetDevice().ApplicationUUID != i.GetDevice().ApplicationUUID {
			t.Error("expected" + d.GetDevice().ApplicationUUID)
		}
		if d.GetDevice().LocalUUID != i.GetDevice().LocalUUID {
			t.Error("expected" + d.GetDevice().LocalUUID)
		}
		if d.GetDevice().Radio != i.GetDevice().Radio {
			t.Errorf(`expected %s`, d.GetDevice().Radio)
		}
	}

	r = radio.Create(radio.WIFI)
	d.GetDevice().ApplicationUUID = "_application"
	d.GetDevice().LocalUUID = "_local"
	d.GetDevice().Radio = r

	if err := db.Update(key, d); err != nil {
		t.Error("Update device failed")
		t.Error(err)
	} else if i, err := db.Get(key, device.NRF51822, r); err != nil {
		t.Error("Update device failed")
		t.Error(err)
	} else {
		if d.GetDevice().ApplicationUUID != i.GetDevice().ApplicationUUID {
			t.Error("expected" + d.GetDevice().ApplicationUUID)
		}
		if d.GetDevice().LocalUUID != i.GetDevice().LocalUUID {
			t.Error("expected" + d.GetDevice().LocalUUID)
		}
		if d.GetDevice().Radio != i.GetDevice().Radio {
			t.Errorf(`expected %s`, d.GetDevice().Radio)
		}
	}
}

func TestQuery(t *testing.T) {
	dir := "/tmp/MyDatabase"
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	db := &Database{Directory: dir}
	if err := db.Start(); err != nil {
		t.Error("Start db failed")
		t.Error(err)
	}
	defer db.Stop()

	d := device.Create(device.NRF51822)
	r := radio.Create(radio.BLUETOOTH)
	d.GetDevice().ApplicationUUID = "application_"
	d.GetDevice().LocalUUID = "local"
	d.GetDevice().Radio = r

	key1, err := db.Create(d)
	if err != nil {
		t.Error("Query device failed")
		t.Error(err)
	} else if i, err := db.Get(key1, device.NRF51822, r); err != nil {
		t.Error("Query device failed")
		t.Error(err)
	} else {
		if d.GetDevice().ApplicationUUID != i.GetDevice().ApplicationUUID {
			t.Error("expected" + d.GetDevice().ApplicationUUID)
		}
		if d.GetDevice().LocalUUID != i.GetDevice().LocalUUID {
			t.Error("expected" + d.GetDevice().LocalUUID)
		}
		if d.GetDevice().Radio != i.GetDevice().Radio {
			t.Errorf(`expected %s`, d.GetDevice().Radio)
		}
	}

	d = device.Create(device.NRF51822)
	r = radio.Create(radio.BLUETOOTH)
	d.GetDevice().ApplicationUUID = "application"
	d.GetDevice().LocalUUID = "local_"
	d.GetDevice().Radio = r

	key2, err := db.Create(d)
	if err != nil {
		t.Error("Query device failed")
		t.Error(err)
	} else if i, err := db.Get(key2, device.NRF51822, r); err != nil {
		t.Error("Query device failed")
		t.Error(err)
	} else {
		if d.GetDevice().ApplicationUUID != i.GetDevice().ApplicationUUID {
			t.Error("expected" + d.GetDevice().ApplicationUUID)
		}
		if d.GetDevice().LocalUUID != i.GetDevice().LocalUUID {
			t.Error("expected" + d.GetDevice().LocalUUID)
		}
		if d.GetDevice().Radio != i.GetDevice().Radio {
			t.Errorf(`expected %s`, d.GetDevice().Radio)
		}
	}

	if devices, err := db.Query("ApplicationUUID", "application_", device.NRF51822, r); err != nil {
		t.Error("Query device failed")
		t.Error(err)
	} else {
		if devices[key1].GetDevice().ApplicationUUID != "application_" {
			t.Error("expected" + "application_")
		}
		if devices[key1].GetDevice().LocalUUID != "local" {
			t.Error("expected" + "local")
		}
	}

	if devices, err := db.Query("LocalUUID", "local_", device.NRF51822, r); err != nil {
		t.Error("Query device failed")
		t.Error(err)
	} else {
		if devices[key2].GetDevice().ApplicationUUID != "application" {
			t.Error("expected" + "application")
		}
		if devices[key2].GetDevice().LocalUUID != "local_" {
			t.Error("expected" + "local_")
		}
	}
}
