package device

import (
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/radio"
)

// Type contains the micro and radio that make up a device type
type Type struct {
	Micro micro.Type `mapstructure:"micro" structs:"micro"`
	Radio radio.Type `mapstructure:"radio" structs:"radio"`
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
	Type         `mapstructure:"type" structs:"type"`
	LocalUUID    string    `mapstructure:"localUUID" structs:"localUUID"`
	ResinUUID    string    `mapstructure:"resinUUID" structs:"resinUUID"`
	Name         string    `mapstructure:"name" structs:"name"`
	AppUUID      int       `mapstructure:"applicationUUID" structs:"applicationUUID"`
	AppName      string    `mapstructure:"applicationName" structs:"applicationName"`
	DatabaseUUID int       `mapstructure:"databaseUUID" structs:"databaseUUID"`
	Commit       string    `mapstructure:"commit" structs:"commit"`
	TargetCommit string    `mapstructure:"targetCommit" structs:"-"`
	LastSeen     time.Time `mapstructure:"lastSeen" structs:"lastSeen"`
	State        State     `mapstructure:"state" structs:"state"`
	Progress     float32   `mapstructure:"progress" structs:"progress"`
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
			"Application name: %s, "+
			"Database UUID: %d, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Last seen: %s, "+
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
