package device

import (
	"encoding/json"
	"fmt"

	"github.com/resin-io/edge-node-manager/board"
	"github.com/resin-io/edge-node-manager/database"
	"github.com/resin-io/edge-node-manager/device/status"
	"github.com/resin-io/edge-node-manager/supervisor"
)

type Device struct {
	Board           board.Interface `json:"-"`
	Name            string          `json:"name"`
	BoardType       board.Type      `json:"boardType"`
	LocalUUID       string          `json:"localUUID"`
	ResinUUID       string          `json:"resinUUID"`
	ApplicationUUID int             `json:"applicationUUID"`
	ApplicationName string          `json:"applicationName"`
	Commit          string          `json:"commit"`
	TargetCommit    string          `json:"targetCommit"`
	Status          status.Status   `json:"status"`
	Progress        float32         `json:"progress"`
	Config          interface{}     `json:"config"`
	Environment     interface{}     `json:"environment"`
	RestartFlag     bool            `json:"restartFlag"`
	DeleteFlag      bool            `json:"deleteFlag"`
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
			"Progress: %2.2f, "+
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
		d.Config,
		d.Environment)
}

func Create(boardType board.Type, name, localUUID, resinUUID string, applicationUUID int, applicationName, targetCommit string, config, environment interface{}) (*Device, error) {
	d := &Device{
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
		Config:          config,
		Environment:     environment,
	}

	return d, d.putDevice()
}

func (d *Device) putDevice() error {
	bytes, err := d.Marshall()
	if err != nil {
		return err
	}

	return database.PutDevice(d.ApplicationUUID, d.LocalUUID, d.ResinUUID, bytes)
}

func (d *Device) Marshall() ([]byte, error) {
	return json.Marshal(d)
}

func Unmarshall(bytes []byte) (*Device, error) {
	var d *Device
	if err := json.Unmarshal(bytes, &d); err != nil {
		return nil, err
	}

	d.Board = board.Create(d.BoardType, d.LocalUUID)

	return d, nil
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
