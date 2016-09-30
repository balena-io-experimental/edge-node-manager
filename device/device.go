package device

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/radio"
	"github.com/josephroberts/edge-node-manager/supervisor"
)

// Type contains the micro and radio that make up a device type
type Type struct {
	Micro micro.Type `json:"micro"`
	Radio radio.Type `json:"radio"`
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
	Type            `json:"type"`
	DeviceType      string      `json:"device_type"` // TODO: Need to add below
	Note            string      `json:"note"`        // TODO: Need to add below
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
	Config          interface{} `json:"config"`      // TODO: Need to add below
	Environment     interface{} `json:"environment"` // TODO: Need to add below
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
			"Identify flag: %t",
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
		d.IdentifyFlag)
}

// Update updates a specific device
func (d Device) Update(path string) error {
	d.SetState(UPDATING)
	err := d.Cast().Update(path)
	d.SetState(ONLINE)
	return err
}

// Online checks if a specific device is online
func (d Device) Online() (bool, error) {
	return d.Cast().Online()
}

// New creates a new device and puts it into the database
func New(deviceType Type, localUUID, UUID, name string, applicationUUID int, applicationName, targetCommit string) error {
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
		State:           ONLINE,
		Progress:        0.0,
		RestartFlag:     false,
		IdentifyFlag:    false,
	}

	buffer, err := json.Marshal(newDevice)
	if err != nil {
		return err
	}

	return database.PutDevice(newDevice.ApplicationUUID, newDevice.UUID, buffer)
}

// PutAll puts all devices for a specific application into the database
func PutAll(applicationUUID int, devices map[string]*Device) error {
	buffer := make(map[string][]byte)
	for _, value := range devices {
		bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		buffer[value.UUID] = bytes
	}

	return database.PutDevices(applicationUUID, buffer)
}

// GetAll gets all devices for a specific application
func GetAll(applicationUUID int) (map[string]*Device, error) {
	buffer, err := database.GetDevices(applicationUUID)
	if err != nil {
		return nil, err
	}

	devices := make(map[string]*Device)
	for _, value := range buffer {
		var device Device
		if err = json.Unmarshal(value, &device); err != nil {
			return nil, err
		}

		devices[device.LocalUUID] = &device
	}

	return devices, nil
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

	online := false
	if d.State == ONLINE {
		d.LastSeen = time.Now()
		online = true
	}

	// TODO: handle error
	supervisor.DependantDeviceInfoUpdate(d.UUID, (string)(d.State), online)
}
