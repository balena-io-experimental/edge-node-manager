package device

import (
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/board"
	"github.com/resin-io/edge-node-manager/database"
	"github.com/resin-io/edge-node-manager/device/hook"
	"github.com/resin-io/edge-node-manager/device/status"
	"github.com/resin-io/edge-node-manager/supervisor"
)

type Device struct {
	Log               *logrus.Logger  `json:"-"`
	Board             board.Interface `json:"-"`
	Name              string          `json:"name"`
	BoardType         board.Type      `json:"boardType"`
	LocalUUID         string          `json:"localUUID"`
	ResinUUID         string          `json:"resinUUID"`
	ApplicationUUID   int             `json:"applicationUUID"`
	ApplicationName   string          `json:"applicationName"`
	Commit            string          `json:"commit"`
	TargetCommit      string          `json:"targetCommit"`
	Status            status.Status   `json:"status"`
	Progress          float32         `json:"progress"`
	Config            interface{}     `json:"config"`
	TargetConfig      interface{}     `json:"targetConfig"`
	Environment       interface{}     `json:"environment"`
	TargetEnvironment interface{}     `json:"targetEnvironment"`
	RestartFlag       bool            `json:"restartFlag"`
	DeleteFlag        bool            `json:"deleteFlag"`
	statusFlag        bool            // Used to ensure the is_online is always sent first time after a restart
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
			"Target config: %v, "+
			"Environment: %v, "+
			"Target environment: %v",
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
		d.TargetConfig,
		d.Environment,
		d.TargetEnvironment)
}

func Create(boardType board.Type, name, localUUID, resinUUID string, applicationUUID int, applicationName, targetCommit string) (*Device, error) {
	log := hook.Create(resinUUID)

	board, err := board.Create(boardType, localUUID, log)
	if err != nil {
		return nil, err
	}

	d := &Device{
		Log:             log,
		Board:           board,
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
	}

	return d, d.putDevice()
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

	// Send is_online if the status has changed or its the first time after a restart
	if oldOnline != newOnline || !d.statusFlag {
		d.statusFlag = true
		return supervisor.DependantDeviceInfoUpdateWithOnlineState(d.ResinUUID, (string)(d.Status), d.Commit, newOnline)
	}
	return supervisor.DependantDeviceInfoUpdateWithoutOnlineState(d.ResinUUID, (string)(d.Status), d.Commit)
}

func (d *Device) Marshall() ([]byte, error) {
	return json.Marshal(d)
}

func Unmarshall(bytes []byte) (*Device, error) {
	var d *Device
	if err := json.Unmarshal(bytes, &d); err != nil {
		return nil, err
	}

	log := hook.Create(d.ResinUUID)

	board, err := board.Create(d.BoardType, d.LocalUUID, log)
	if err != nil {
		return nil, err
	}

	d.Board = board
	d.Log = log

	return d, nil
}

// Sync device with resin to ensure we have the latest values for:
// - Device name
// - Device target config
// - Device target environment
func (d *Device) Sync() []error {
	bytes, errs := supervisor.DependantDeviceInfo(d.ResinUUID)
	if errs != nil {
		return errs
	}

	fmt.Println("YO")

	buffer, err := Unmarshall(bytes)
	if err != nil {
		return []error{err}
	}

	fmt.Println("YO1")

	fmt.Println(buffer)

	fmt.Println("YO2")

	d.Name = buffer.Name
	d.TargetConfig = buffer.TargetConfig
	d.TargetEnvironment = buffer.TargetEnvironment

	fmt.Println(d)

	fmt.Println("YO4")

	return nil
}

func (d *Device) putDevice() error {
	bytes, err := d.Marshall()
	if err != nil {
		return err
	}

	return database.PutDevice(d.ApplicationUUID, d.LocalUUID, d.ResinUUID, bytes)
}
