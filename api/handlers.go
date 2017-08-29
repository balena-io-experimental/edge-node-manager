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
	type dependentDeviceUpdate struct {
		Commit      string      `json:"commit"`
		Environment interface{} `json:"environment"`
	}

	decoder := json.NewDecoder(r.Body)
	var content dependentDeviceUpdate
	if err := decoder.Decode(&content); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to decode Dependent device update hook")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := setField(r, "TargetCommit", content.Commit); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func DependentDeviceDelete(w http.ResponseWriter, r *http.Request) {
	if err := setField(r, "Delete", true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DependentDeviceRestart(w http.ResponseWriter, r *http.Request) {
	if err := setField(r, "Restart", true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DependentDevicesQuery(w http.ResponseWriter, r *http.Request) {
	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var d []device.Device
	if err := db.All(&d); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to find devices in database")
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
	vars := mux.Vars(r)
	UUID := vars["uuid"]

	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var d device.Device
	if err := db.Select(
		q.Or(
			q.Eq("LocalUUID", UUID),
			q.Eq("ResinUUID", UUID),
		),
	).First(&d); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"UUID":  UUID,
		}).Error("Unable to find device in database")
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
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&content); err != nil {
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

func setField(r *http.Request, key string, value interface{}) error {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		return err
	}
	defer db.Close()

	var d device.Device
	if err := db.One("ResinUUID", deviceUUID, &d); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"UUID":  deviceUUID,
		}).Error("Unable to find device in database")
		return err
	}

	switch key {
	case "TargetCommit":
		d.TargetCommit = value.(string)
	case "Delete":
		d.DeleteFlag = value.(bool)
	case "Restart":
		d.RestartFlag = value.(bool)
	default:
		log.WithFields(log.Fields{
			"Error": err,
			"UUID":  deviceUUID,
			"Key":   key,
			"value": value,
		}).Error("Unable to set field")
		return err
	}

	if err := db.Update(&d); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"UUID":  deviceUUID,
		}).Error("Unable to update device in database")
		return err
	}

	log.WithFields(log.Fields{
		"UUID":  deviceUUID,
		"Key":   key,
		"value": value,
	}).Debug("Dependent device field updated")

	return nil
}
