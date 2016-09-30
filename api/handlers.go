package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/database"

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
	}

	log.WithFields(log.Fields{
		"UUID":   deviceUUID,
		"Commit": content.Commit,
		"Env":    content.Environment,
	}).Debug("Dependant device update hook")

	applicationUUID, err := database.GetDeviceMapping(deviceUUID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to get device mapping")
		return
	}

	if err = database.PutDeviceField(applicationUUID, deviceUUID, "targetCommit", ([]byte)(content.Commit)); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to put target commit")
		return
	}

	log.WithFields(log.Fields{
		"App":    applicationUUID,
		"Device": deviceUUID,
		"Target": content.Commit,
	}).Debug("Set device target commit")

	application.List[applicationUUID].TargetCommit = content.Commit

	log.WithFields(log.Fields{
		"App":    applicationUUID,
		"Target": content.Commit,
	}).Debug("Set app target commit")
}

// DependantDeviceRestart puts the restart flag for a specific device
func DependantDeviceRestart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

	log.WithFields(log.Fields{
		"UUID": deviceUUID,
	}).Debug("Dependant device restart hook")

	applicationUUID, err := database.GetDeviceMapping(deviceUUID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to get device mapping")
		return
	}

	if err = database.PutDeviceField(applicationUUID, deviceUUID, "restartFlag", ([]byte)(strconv.FormatBool(true))); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to put restart flag")
		return
	}
}

// DependantDeviceIdentify puts the identify flag for a specific device
func DependantDeviceIdentify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

	log.WithFields(log.Fields{
		"UUID": deviceUUID,
	}).Debug("Dependant device identify hook")

	applicationUUID, err := database.GetDeviceMapping(deviceUUID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to get device mapping")
		return
	}

	if err = database.PutDeviceField(applicationUUID, deviceUUID, "identifyFlag", ([]byte)(strconv.FormatBool(true))); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to put identify flag")
		return
	}

	fmt.Fprintln(w, "Dependant Device Identify")
}
