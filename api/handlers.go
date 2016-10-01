package api

import (
	"encoding/json"
	"net/http"

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

	applicationUUID, localUUID, err := database.GetDeviceMapping(deviceUUID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to get device mapping")
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
}

// TODO: DependantDeviceRestart puts the restart flag for a specific device
func DependantDeviceRestart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

	log.WithFields(log.Fields{
		"UUID": deviceUUID,
	}).Debug("Dependant device restart hook")

	// applicationUUID, err := database.GetDeviceMapping(deviceUUID)
	// if err != nil {
	// 	log.WithFields(log.Fields{
	// 		"Error": err,
	// 	}).Error("Unable to get device mapping")
	// 	return
	// }

	// if err = database.PutDeviceField(applicationUUID, deviceUUID, "restartFlag", ([]byte)(strconv.FormatBool(true))); err != nil {
	// 	log.WithFields(log.Fields{
	// 		"Error": err,
	// 	}).Error("Unable to put restart flag")
	// 	return
	// }
}

// TODO: DependantDeviceIdentify puts the identify flag for a specific device
func DependantDeviceIdentify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceUUID := vars["uuid"]

	log.WithFields(log.Fields{
		"UUID": deviceUUID,
	}).Debug("Dependant device identify hook")

	// applicationUUID, err := database.GetDeviceMapping(deviceUUID)
	// if err != nil {
	// 	log.WithFields(log.Fields{
	// 		"Error": err,
	// 	}).Error("Unable to get device mapping")
	// 	return
	// }

	// if err = database.PutDeviceField(applicationUUID, deviceUUID, "identifyFlag", ([]byte)(strconv.FormatBool(true))); err != nil {
	// 	log.WithFields(log.Fields{
	// 		"Error": err,
	// 	}).Error("Unable to put identify flag")
	// 	return
	// }

	// fmt.Fprintln(w, "Dependant Device Identify")
}
