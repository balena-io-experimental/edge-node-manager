package device

import (
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/radio"
)

// Type contains the micro and radio that make up a device type
type Type struct {
	Micro micro.Type `json:"Micro"`
	Radio radio.Type `json:"Radio"`
}

// State defines the device states
type State string

const (
	UPDATING    State = "Updating"
	ONLINE            = "Online"
	OFFLINE           = "Offline"
	DOWNLOADING       = "Downloading"
)

// Device contains all the variables needed to define a device
type Device struct {
	Type         `json:"type"`
	LocalUUID    string    `json:"localUUID" structs:"localUUID"`
	ResinUUID    string    `json:"resinUUID" structs:"resinUUID"`
	Name         string    `json:"name" structs:"name"`
	AppUUID      int       `json:"applicationUUID" structs:"applicationUUID"`
	AppName      string    `json:"applicationName" structs:"applicationName"`
	DatabaseUUID int       `json:"databaseUUID" structs:"databaseUUID"`
	Commit       string    `json:"commit" structs:"commit"`
	TargetCommit string    `json:"targetCommit" structs:"-"`
	LastSeen     time.Time `json:"lastSeen" structs:"lastSeen"`
	State        State     `json:"state" structs:"state"`
	Progress     float32   `json:"progress" structs:"progress"`
}

// Interface defines the common functions a device must implement
type Interface interface {
	String() string
	Update(commit, directory string) error
	Online() (bool, error)
	Identify() error
	Restart() error
}

func (d Device) String() string {
	return fmt.Sprintf(
		"Micro type: %s, "+
			"Radio type: %s, "+
			"Local UUID: %s, "+
			"Resin UUID: %s, "+
			"Name: %s, "+
			"Application UUID: %d, "+
			"Application Name: %s, "+
			"Database UUID: %d, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"LastSeen: %s, "+
			"State: %s, "+
			"Progress: %2.2f",
		d.Type.Micro,
		d.Type.Radio,
		d.LocalUUID,
		d.ResinUUID,
		d.Name,
		d.AppUUID,
		d.AppName,
		d.DatabaseUUID,
		d.Commit,
		d.TargetCommit,
		d.LastSeen,
		d.State,
		d.Progress)
}

// New creates a new device
func New(deviceType Type, localUUID, resinUUID, name string, appUUID int, appName, targetCommit string) *Device {
	newDevice := &Device{
		Type:         deviceType,
		LocalUUID:    localUUID,
		ResinUUID:    resinUUID,
		Name:         name,
		AppUUID:      appUUID,
		AppName:      appName,
		DatabaseUUID: 0,
		Commit:       "",
		TargetCommit: targetCommit,
		LastSeen:     time.Now(),
		State:        ONLINE,
		Progress:     0.0,
	}

	return newDevice
}

// Cast converts the base device type to its actual type
func (d Device) Cast() Interface {
	switch d.Type.Micro {
	case micro.NRF51822:
		return (Nrf51822)(d)
	case micro.ESP8266:
		return (Esp8266)(d)
	}

	return nil
}
