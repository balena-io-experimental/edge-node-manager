package api

import (
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
	targetCommit := vars["commit"]

	applicationUUID, err := database.GetDeviceMapping(deviceUUID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to get device mapping")
		return
	}

	if err = database.PutDeviceField(applicationUUID, deviceUUID, "targetCommit", ([]byte)(targetCommit)); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to put target commit")
		return
	}

	application.List[applicationUUID].TargetCommit = targetCommit

	fmt.Fprintln(w, "Dependant Device Update")
}

// DependantDeviceRestart puts the restart flag for a specific device
func DependantDeviceRestart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

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

	fmt.Fprintln(w, "Dependant Device Restart")
}

// DependantDeviceIdentify puts the identify flag for a specific device
func DependantDeviceIdentify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

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
