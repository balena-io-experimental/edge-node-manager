package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/database"

	log "github.com/Sirupsen/logrus"
)

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

	database.PutDeviceField(applicationUUID, deviceUUID, "targetCommit", ([]byte)(targetCommit))

	application.List[applicationUUID].TargetCommit = targetCommit
}

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

	database.PutDeviceField(applicationUUID, deviceUUID, "restartFlag", ([]byte)(strconv.FormatBool(true)))
}

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

	database.PutDeviceField(applicationUUID, deviceUUID, "identifyFlag", ([]byte)(strconv.FormatBool(true)))
}
