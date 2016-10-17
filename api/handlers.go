package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/process"

	log "github.com/Sirupsen/logrus"
)

// DependantDeviceUpdate puts the target commit for a specific device and its parent application
func DependantDeviceUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

	type dependantDeviceUpdate struct {
		Commit      string      `json:"commit"`
		Environment interface{} `json:"environment"`
	}

	decoder := json.NewDecoder(r.Body)
	var content dependantDeviceUpdate
	if err := decoder.Decode(&content); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to decode Dependant device update hook")
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

	log.WithFields(log.Fields{
		"ApplicationUUID": applicationUUID,
		"DeviceUUID":      deviceUUID,
		"LocalUUID":       localUUID,
		"Target commit":   content.Commit,
	}).Debug("Dependant device update hook")

	application.List[applicationUUID].TargetCommit = content.Commit
	application.List[applicationUUID].Devices[localUUID].TargetCommit = content.Commit

	w.WriteHeader(http.StatusOK)
}

// DependantDeviceDelete deletes a specific device
func DependantDeviceDelete(w http.ResponseWriter, r *http.Request) {
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

	log.WithFields(log.Fields{
		"ApplicationUUID": applicationUUID,
		"DeviceUUID":      deviceUUID,
		"LocalUUID":       localUUID,
	}).Debug("Dependant device delete hook")

	delete(application.List[applicationUUID].Devices, localUUID)
	if err := database.DeleteDevice(applicationUUID, deviceUUID); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to delete device")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DependantDeviceRestart puts the restart flag for a specific device
func DependantDeviceRestart(w http.ResponseWriter, r *http.Request) {
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

	log.WithFields(log.Fields{
		"UUID": deviceUUID,
	}).Debug("Dependant device restart hook")

	application.List[applicationUUID].Devices[localUUID].RestartFlag = true

	w.WriteHeader(http.StatusOK)
}

// PauseTarget sets the process pause target flag
func PauseTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	target, err := strconv.ParseBool(vars["state"])
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to parse state field")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"Pause target": target,
	}).Debug("Set pause hook")

	process.PauseTarget = target

	w.WriteHeader(http.StatusOK)
}

// PauseState gets the process pause state flag
func PauseState(w http.ResponseWriter, r *http.Request) {
	type pauseState struct {
		State bool `json:"state"`
	}

	content := &pauseState{
		State: process.PauseState,
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to parse state field")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"Pause state": process.PauseState,
	}).Debug("Get pause hook")

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	w.WriteHeader(http.StatusOK)
}
