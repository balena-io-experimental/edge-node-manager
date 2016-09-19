package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/josephroberts/edge-node-manager/database"
)

func DependantDeviceUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resinUUID := vars["uuid"]
	commit := vars["commit"]
	fmt.Fprintln(w, "Dependant Device Update:", resinUUID)

	database.SetTargetCommit(resinUUID, commit)
}

func DependantDeviceRestart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resinUUID := vars["ResinUUID"]
	fmt.Fprintln(w, "Dependant Device Restart", resinUUID)
	// TODO
}

func DependantDeviceIdentify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resinUUID := vars["ResinUUID"]
	fmt.Fprintln(w, "Dependant Device Identify", resinUUID)
	// TODO
}
