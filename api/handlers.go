package api

import (
	"encoding/json"
	"net/http"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gorilla/mux"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/device"
	"github.com/resin-io/edge-node-manager/process"
	"github.com/resin-io/edge-node-manager/process/status"

	log "github.com/Sirupsen/logrus"
)

func DependentDeviceUpdate(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	type deviceTarget struct {
		Commit      string                 `json:"commit"`
		Environment map[string]interface{} `json:"environment"`
		Config      map[string]interface{} `json:"config"`
	}

	var t deviceTarget
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to decode Dependent device update hook")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	d, err := loadDevice(db, uuid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	d.TargetCommit = t.Commit
	d.TargetEnvironment = t.Environment
	d.TargetConfig = t.Config

	if err := updateDevice(db, d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"UUID": uuid,
	}).Debug("Dependent device field updated")

	w.WriteHeader(http.StatusAccepted)
}

func DependentDeviceDelete(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	d, err := loadDevice(db, uuid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	d.DeleteFlag = true

	if err := updateDevice(db, d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"UUID": uuid,
	}).Debug("Dependent device field updated")

	w.WriteHeader(http.StatusOK)
}

func DependentDeviceRestart(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	d, err := loadDevice(db, uuid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	d.RestartFlag = true

	if err := updateDevice(db, d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"UUID": uuid,
	}).Debug("Dependent device field updated")

	w.WriteHeader(http.StatusOK)
}

func DependentDevicesQuery(w http.ResponseWriter, r *http.Request) {
	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	d, err := loadDevices(db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(d)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to encode devices")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if written, err := w.Write(bytes); (err != nil) || (written != len(bytes)) {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to write response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug("Get dependent device")
}

func DependentDeviceQuery(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	d, err := loadDevice(db, uuid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(d)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to encode device")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if written, err := w.Write(bytes); (err != nil) || (written != len(bytes)) {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to write response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"Device": d,
	}).Debug("Get dependent device")
}

func SetStatus(w http.ResponseWriter, r *http.Request) {
	type s struct {
		TargetStatus status.Status `json:"targetStatus"`
	}

	var content *s
	if err := json.NewDecoder(r.Body).Decode(&content); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to decode status hook")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	process.TargetStatus = content.TargetStatus

	w.WriteHeader(http.StatusOK)

	log.WithFields(log.Fields{
		"Target status": process.TargetStatus,
	}).Debug("Set status")
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	type s struct {
		CurrentStatus status.Status `json:"currentStatus"`
		TargetStatus  status.Status `json:"targetStatus"`
	}

	content := &s{
		CurrentStatus: process.CurrentStatus,
		TargetStatus:  process.TargetStatus,
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to encode status hook")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if written, err := w.Write(bytes); (err != nil) || (written != len(bytes)) {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to write response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"Target status": process.TargetStatus,
		"Curent status": process.CurrentStatus,
	}).Debug("Get status")
}

func openDB() (*storm.DB, error) {
	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to open database")
		return nil, err
	}

	return db, nil
}

func loadDevice(db *storm.DB, uuid string) (device.Device, error) {
	var d device.Device
	if err := db.Select(
		q.Or(
			q.Eq("LocalUUID", uuid),
			q.Eq("ResinUUID", uuid),
		),
	).First(&d); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"UUID":  uuid,
		}).Error("Unable to find device in database")
		return d, err
	}

	return d, nil
}

func loadDevices(db *storm.DB) ([]device.Device, error) {
	var d []device.Device
	if err := db.All(&d); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to find devices in database")
		return d, err
	}

	return d, nil
}

func updateDevice(db *storm.DB, d device.Device) error {
	if err := db.Update(&d); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to update device in database")
		return err
	}

	return nil
}
