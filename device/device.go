package device

import (
	"encoding/json"
	"fmt"

	"github.com/resin-io/edge-node-manager/board"
	"github.com/resin-io/edge-node-manager/board/esp8266"
	"github.com/resin-io/edge-node-manager/board/microbit"
	"github.com/resin-io/edge-node-manager/board/nrf51822dk"
	"github.com/resin-io/edge-node-manager/device/hook"
	"github.com/resin-io/edge-node-manager/device/status"
	"github.com/resin-io/edge-node-manager/micro/nrf51822"
	"github.com/resin-io/edge-node-manager/supervisor"
)

type Device struct {
	Board             board.Interface        `json:"-"`
	ApplicationUUID   int                    `storm:"index"`
	BoardType         board.Type             `storm:"index"`
	Name              string                 `storm:"index"`
	LocalUUID         string                 `storm:"index"`
	ResinUUID         string                 `storm:"id,unique,index"`
	Commit            string                 `storm:"index"`
	TargetCommit      string                 `storm:"index"`
	Status            status.Status          `storm:"index"`
	Config            map[string]interface{} `storm:"index"`
	TargetConfig      map[string]interface{} `storm:"index"`
	Environment       map[string]interface{} `storm:"index"`
	TargetEnvironment map[string]interface{} `storm:"index"`
	RestartFlag       bool                   `storm:"index"`
	DeleteFlag        bool                   `storm:"index"`
}

func (d Device) String() string {
	return fmt.Sprintf(
		"Application UUID: %d, "+
			"Board type: %s, "+
			"Name: %s, "+
			"Local UUID: %s, "+
			"Resin UUID: %s, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Status: %s, "+
			"Config: %v, "+
			"Target config: %v, "+
			"Environment: %v, "+
			"Target environment: %v, "+
			"Restart: %t, "+
			"Delete: %t",
		d.ApplicationUUID,
		d.BoardType,
		d.Name,
		d.LocalUUID,
		d.ResinUUID,
		d.Commit,
		d.TargetCommit,
		d.Status,
		d.Config,
		d.TargetConfig,
		d.Environment,
		d.TargetEnvironment,
		d.RestartFlag,
		d.DeleteFlag)
}

func New(applicationUUID int, boardType board.Type, name, localUUID, resinUUID string) Device {
	return Device{
		ApplicationUUID:   applicationUUID,
		BoardType:         boardType,
		Name:              name,
		LocalUUID:         localUUID,
		ResinUUID:         resinUUID,
		Commit:            "",
		TargetCommit:      "",
		Status:            status.OFFLINE,
		Config:            make(map[string]interface{}),
		TargetConfig:      make(map[string]interface{}),
		Environment:       make(map[string]interface{}),
		TargetEnvironment: make(map[string]interface{}),
	}
}

func (d *Device) PopulateBoard() error {
	log := hook.Create(d.ResinUUID)

	switch d.BoardType {
	case board.MICROBIT:
		d.Board = microbit.Microbit{
			Log: log,
			Micro: nrf51822.Nrf51822{
				Log:                 log,
				LocalUUID:           d.LocalUUID,
				Firmware:            nrf51822.FIRMWARE{},
				NotificationChannel: make(chan []byte),
			},
		}
	case board.NRF51822DK:
		d.Board = nrf51822dk.Nrf51822dk{
			Log: log,
			Micro: nrf51822.Nrf51822{
				Log:                 log,
				LocalUUID:           d.LocalUUID,
				Firmware:            nrf51822.FIRMWARE{},
				NotificationChannel: make(chan []byte),
			},
		}
	case board.ESP8266:
		d.Board = esp8266.Esp8266{
			Log:       log,
			LocalUUID: d.LocalUUID,
		}
	default:
		return fmt.Errorf("Unsupported board type")
	}

	return nil
}

// Sync device with resin to ensure we have the latest values for:
// - Device name
func (d *Device) Sync() []error {
	bytes, errs := supervisor.DependentDeviceInfo(d.ResinUUID)
	if errs != nil {
		return errs
	}

	var temp Device
	if err := json.Unmarshal(bytes, &temp); err != nil {
		// Ignore the error here as it means the device we are trying
		// to sync has been deleted
		return nil
	}

	d.Name = temp.Name

	return nil
}
