package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func DependantDeviceUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resinUUID := vars["ResinUUID"]
	fmt.Fprintln(w, "Dependant Device Update:", resinUUID)
	// TODO
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
