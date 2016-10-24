package device

import (
	"fmt"

	"github.com/josephroberts/edge-node-manager/board"
	"github.com/josephroberts/edge-node-manager/device/status"
	"github.com/josephroberts/edge-node-manager/supervisor"
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

func Create(boardType board.Type, name, localUUID, resinUUID string, applicationUUID int, applicationName, targetCommit string, config, environment interface{}) *Device {
	return &Device{
		Board:           board.Create(boardType, localUUID),
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

// Only set the is_online field if the device is_online state has changed
func (d *Device) SetStatus(newStatus status.Status) []error {
	oldOnline := true
	if d.Status == status.OFFLINE {
		oldOnline = false
	}

	d.Status = newStatus

	newOnline := true
	if d.Status == status.OFFLINE {
		newOnline = false
	}

	if oldOnline != newOnline {
		return supervisor.DependantDeviceInfoUpdateWithOnlineState(d.ResinUUID, (string)(d.Status), d.Commit, newOnline)
	}
	return supervisor.DependantDeviceInfoUpdateWithoutOnlineState(d.ResinUUID, (string)(d.Status), d.Commit)
}
