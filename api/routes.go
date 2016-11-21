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
		"DependentDeviceUpdate",
		"PUT",
		"/v1/devices/{uuid}",
		DependentDeviceUpdate,
	},
	Route{
		"DependentDeviceDelete",
		"DELETE",
		"/v1/devices/{uuid}",
		DependentDeviceDelete,
	},
	Route{
		"DependentDeviceRestart",
		"PUT",
		"/v1/devices/{uuid}/restart",
		DependentDeviceRestart,
	},
	Route{
		"SetStatus",
		"PUT",
		"/v1/enm/status",
		SetStatus,
	},
	Route{
		"GetStatus",
		"GET",
		"/v1/enm/status",
		GetStatus,
	},
	Route{
		"Pending",
		"GET",
		"/v1/enm/pending",
		Pending,
	},
}
