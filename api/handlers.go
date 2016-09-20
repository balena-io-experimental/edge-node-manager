package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/josephroberts/edge-node-manager/database"
)

func DependantDeviceUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resinUUID := vars["uuid"]
	commit := vars["commit"]

	device := database.LoadDevice(resinUUID)
	database.UpdateField(device.DatabaseUUID, "targetCommit", commit)
	applications.List[device.AppName] = commit
}

func DependantDeviceRestart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resinUUID := vars["ResinUUID"]

	device := database.LoadDevice(resinUUID)
	database.UpdateField(device.DatabaseUUID, "restart", true)
}

func DependantDeviceIdentify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resinUUID := vars["ResinUUID"]

	device := database.LoadDevice(resinUUID)
	database.UpdateField(device.DatabaseUUID, "identify", true)
}
