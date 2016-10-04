package device

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/radio"
)

// Type contains the micro and radio that make up a device type
type Type struct {
	Micro micro.Type `json:"micro"`
	Radio radio.Type `json:"radio"`
}

// State defines the device states
type State string

const (
	DOWNLOADING State = "Downloading"
	INSTALLING        = "Installing"
	STARTING          = "Starting"
	STOPPING          = "Stopping"
	IDLE              = "Idle"
	OFFLINE           = "Offline"
)

// Device contains all the variables needed to define a device
type Device struct {
	Type            `json:"type"`
	LocalUUID       string      `json:"localUUID"`
	UUID            string      `json:"uuid"`
	Name            string      `json:"name"`
	ApplicationUUID int         `json:"applicationUUID"`
	ApplicationName string      `json:"applicationName"`
	Commit          string      `json:"commit"`
	TargetCommit    string      `json:"targetCommit"`
	LastSeen        time.Time   `json:"lastSeen"`
	State           State       `json:"state"`
	Progress        float32     `json:"progress"`
	RestartFlag     bool        `json:"restartFlag"`
	IdentifyFlag    bool        `json:"identifyFlag"`
	Config          interface{} `json:"config"`
	Environment     interface{} `json:"environment"`
}

// Interface defines the common functions a device must implement
type Interface interface {
	String() string
	Update(path string) error
	Online() (bool, error)
	Identify() error
	Restart() error
}

func (d Device) String() string {
	return fmt.Sprintf(
		"Micro type: %s, "+
			"Radio type: %s, "+
			"Local UUID: %s, "+
			"UUID: %s, "+
			"Name: %s, "+
			"Application UUID: %d, "+
			"Application name: %s, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Last seen: %s, "+
			"State: %s, "+
			"Progress: %2.2f "+
			"Restart flag: %t, "+
			"Identify flag: %t, "+
			"Config: %v, "+
			"Environment: %v",
		d.Type.Micro,
		d.Type.Radio,
		d.LocalUUID,
		d.UUID,
		d.Name,
		d.ApplicationUUID,
		d.ApplicationName,
		d.Commit,
		d.TargetCommit,
		d.LastSeen,
		d.State,
		d.Progress,
		d.RestartFlag,
		d.IdentifyFlag,
		d.Config,
		d.Environment)
}

// Update updates a specific device
func (d Device) Update(path string) error {
	d.SetState(INSTALLING)
	err := d.Cast().Update(path)
	d.SetState(IDLE)
	return err
}

// Online checks if a specific device is online
func (d Device) Online() (bool, error) {
	return d.Cast().Online()
}

// New creates a new device and puts it into the database
func New(deviceType Type, localUUID, UUID, name string, applicationUUID int, applicationName, targetCommit string, config, environment interface{}) error {
	newDevice := &Device{
		Type:            deviceType,
		LocalUUID:       localUUID,
		UUID:            UUID,
		Name:            name,
		ApplicationUUID: applicationUUID,
		ApplicationName: applicationName,
		Commit:          "",
		TargetCommit:    targetCommit,
		LastSeen:        time.Now(),
		State:           IDLE,
		Progress:        0.0,
		RestartFlag:     false,
		IdentifyFlag:    false,
		Config:          config,
		Environment:     environment,
	}

	buffer, err := json.Marshal(newDevice)
	if err != nil {
		return err
	}

	return database.PutDevice(newDevice.ApplicationUUID, newDevice.LocalUUID, newDevice.UUID, buffer)
}

// Get gets a single device for a specific application
func Get(applicationUUID int, UUID string) (*Device, error) {
	buffer, err := database.GetDevice(applicationUUID, UUID)
	if err != nil {
		return nil, err
	}

	var device Device
	if err = json.Unmarshal(buffer, &device); err != nil {
		return nil, err
	}

	return &device, nil
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

// SetState sets the state for a specific device
func (d *Device) SetState(state State) {
	d.State = state

	if d.State != OFFLINE {
		d.LastSeen = time.Now()
	}

	// TODO: Send state update to supervisor
}
