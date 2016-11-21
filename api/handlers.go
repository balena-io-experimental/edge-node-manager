package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/resin-io/edge-node-manager/application"
	"github.com/resin-io/edge-node-manager/database"
	"github.com/resin-io/edge-node-manager/process"
	"github.com/resin-io/edge-node-manager/process/status"

	log "github.com/Sirupsen/logrus"
)

func DependentDeviceUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

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

	applicationUUID, localUUID, err := database.GetDeviceMapping(deviceUUID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to get device mapping")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	application.List[applicationUUID].TargetCommit = content.Commit
	application.List[applicationUUID].Devices[localUUID].TargetCommit = content.Commit

	w.WriteHeader(http.StatusOK)

	log.WithFields(log.Fields{
		"ApplicationUUID": applicationUUID,
		"DeviceUUID":      deviceUUID,
		"LocalUUID":       localUUID,
		"Target commit":   content.Commit,
	}).Debug("Dependent device update hook")
}

func DependentDeviceDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

	applicationUUID, localUUID, err := database.GetDeviceMapping(deviceUUID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to get device mapping")
		// Send back 200 as the device must of already been deleted if we can't find it in the DB
		w.WriteHeader(http.StatusOK)
		return
	}

	application.List[applicationUUID].Devices[localUUID].DeleteFlag = true

	w.WriteHeader(http.StatusOK)

	log.WithFields(log.Fields{
		"ApplicationUUID": applicationUUID,
		"DeviceUUID":      deviceUUID,
		"LocalUUID":       localUUID,
	}).Debug("Dependent device delete hook")
}

func DependentDeviceRestart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

	applicationUUID, localUUID, err := database.GetDeviceMapping(deviceUUID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to get device mapping")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	application.List[applicationUUID].Devices[localUUID].RestartFlag = true

	w.WriteHeader(http.StatusOK)

	log.WithFields(log.Fields{
		"UUID": deviceUUID,
	}).Debug("Dependent device restart hook")
}

func SetStatus(w http.ResponseWriter, r *http.Request) {
	type s struct {
		CurrentStatus status.Status `json:"current"`
		TargetStatus  status.Status `json:"target"`
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
		CurrentStatus status.Status `json:"current"`
		TargetStatus  status.Status `json:"target"`
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
	w.Write(bytes)

	log.WithFields(log.Fields{
		"Target status": process.TargetStatus,
		"Curent status": process.CurrentStatus,
	}).Debug("Get status")
}

func Pending(w http.ResponseWriter, r *http.Request) {
	type p struct {
		Pending bool `json:"pending"`
	}

	pending := process.Pending()

	content := &p{
		Pending: pending,
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to encode pending hook")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)

	log.WithFields(log.Fields{
		"Pending": pending,
	}).Debug("Get status")
}
