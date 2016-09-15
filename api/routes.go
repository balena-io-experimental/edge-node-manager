package api

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"DependantDeviceUpdate",
		"PUT",
		"/v1/devices/{ResinUUID}",
		DependantDeviceUpdate,
	},
	Route{
		"DependantDeviceRestart",
		"PUT",
		"/v1/devices/{ResinUUID}/restart",
		DependantDeviceRestart,
	},
	Route{
		"DependantDeviceIdentify",
		"PUT",
		"/v1/devices/{ResinUUID}/identify",
		DependantDeviceIdentify,
	},
}
