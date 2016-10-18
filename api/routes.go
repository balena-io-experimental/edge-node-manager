package api

import "net/http"

// Route contains all the variables needed to define a route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes holds all the routes assigned to the API
type Routes []Route

var routes = Routes{
	Route{
		"DependantDeviceUpdate",
		"PUT",
		"/v1/devices/{uuid}",
		DependantDeviceUpdate,
	},
	Route{
		"DependantDeviceDelete",
		"DELETE",
		"/v1/devices/{uuid}",
		DependantDeviceDelete,
	},
	Route{
		"DependantDeviceRestart",
		"PUT",
		"/v1/devices/{uuid}/restart",
		DependantDeviceRestart,
	},
	Route{
		"SetTargetState",
		"PUT",
		"/v1/enm/state",
		SetTargetState,
	},
	Route{
		"GetTargetState",
		"GET",
		"/v1/enm/state",
		GetTargetState,
	},
}
