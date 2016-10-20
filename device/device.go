package device

import (
	"fmt"

	"github.com/josephroberts/edge-node-manager/board"
	"github.com/josephroberts/edge-node-manager/device/status"
)

type Device struct {
	Board           board.Interface
	Name            string        `json:"name"`
	BoardType       board.Type    `json:"boardType"`
	LocalUUID       string        `json:"localUUID"`
	ResinUUID       string        `json:"resinUUID"`
	ApplicationUUID int           `json:"applicationUUID"`
	ApplicationName string        `json:"applicationName"`
	Commit          string        `json:"commit"`
	TargetCommit    string        `json:"targetCommit"`
	Status          status.Status `json:"status"`
	Progress        float32       `json:"progress"`
	RestartFlag     bool          `json:"restartFlag"`
	Config          interface{}   `json:"config"`
	Environment     interface{}   `json:"environment"`
}

func (d Device) String() string {
	return fmt.Sprintf(
		"Name: %s, "+
			"Board type: %s, "+
			"Local UUID: %s, "+
			"Resin UUID: %s, "+
			"Application UUID: %d, "+
			"Application name: %s, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Status: %s, "+
			"Progress: %2.2f "+
			"Restart flag: %t, "+
			"Config: %v, "+
			"Environment: %v",
		d.Name,
		d.BoardType,
		d.LocalUUID,
		d.ResinUUID,
		d.ApplicationUUID,
		d.ApplicationName,
		d.Commit,
		d.TargetCommit,
		d.Status,
		d.Progress,
		d.RestartFlag,
		d.Config,
		d.Environment)
}

func Create(boardType board.Type, name, localUUID, resinUUID string, applicationUUID int, applicationName, targetCommit string, config, environment interface{}) Device {
	return Device{
		Board:           board.Create(boardType),
		Name:            name,
		BoardType:       boardType,
		LocalUUID:       localUUID,
		ResinUUID:       resinUUID,
		ApplicationUUID: applicationUUID,
		ApplicationName: applicationName,
		Commit:          "",
		TargetCommit:    targetCommit,
		Status:          status.OFFLINE,
		Progress:        0.0,
		RestartFlag:     false,
		Config:          config,
		Environment:     environment,
	}
}

// // Get gets a single device for a specific application
// func Get(applicationUUID int, UUID string) (*Device, error) {
// 	buffer, err := database.GetDevice(applicationUUID, UUID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var device Device
// 	if err = json.Unmarshal(buffer, &device); err != nil {
// 		return nil, err
// 	}

// 	return &device, nil
// }

// // SetStatus sets the state for a specific device
// // Only set the is_online field if the device is_online state has changed
// func (d *Device) SetStatus(status Status) []error {
// 	// Get the old is_online state
// 	old := true
// 	if d.Status == OFFLINE {
// 		old = false
// 	}

// 	d.Status = status

// 	// Get the new is_online state
// 	new := true
// 	if d.Status == OFFLINE {
// 		new = false
// 	}

// 	// Update the is_online state
// 	if old != new {
// 		return supervisor.DependantDeviceInfoUpdateWithOnlineState(d.UUID, (string)(d.Status), d.Commit, new)
// 	}
// 	// Don't update the is_online state
// 	return supervisor.DependantDeviceInfoUpdateWithoutOnlineState(d.UUID, (string)(d.Status), d.Commit)
// }
