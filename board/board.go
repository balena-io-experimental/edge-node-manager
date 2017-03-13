package board

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/board/cloudjam"
	"github.com/resin-io/edge-node-manager/board/microbit"
	"github.com/resin-io/edge-node-manager/board/nrf51822dk"
	"github.com/resin-io/edge-node-manager/micro/nrf51822"
	"github.com/resin-io/edge-node-manager/micro/stmf401re"
)

type Type string

const (
	MICROBIT   Type = "microbit"
	NRF51822DK      = "nrf51822dk"
	CLOUDJAM        = "cloudjam"
)

type Interface interface {
	Update(path string) error
	Scan(applicationUUID int) (map[string]bool, error)
	Online() (bool, error)
	Restart() error
	Identify() error
	UpdateConfig(interface{}) error
	UpdateEnvironment(interface{}) error
}

func Create(boardType Type, localUUID string, log *logrus.Logger) (Interface, error) {
	switch boardType {
	case MICROBIT:
		return microbit.Microbit{
			Log: log,
			Micro: nrf51822.Nrf51822{
				Log:                 log,
				LocalUUID:           localUUID,
				Firmware:            nrf51822.FIRMWARE{},
				NotificationChannel: make(chan []byte),
			},
		}, nil
	case NRF51822DK:
		return nrf51822dk.Nrf51822dk{
			Log: log,
			Micro: nrf51822.Nrf51822{
				Log:                 log,
				LocalUUID:           localUUID,
				Firmware:            nrf51822.FIRMWARE{},
				NotificationChannel: make(chan []byte),
			},
		}, nil
	case CLOUDJAM:
		return cloudjam.Cloudjam{
			Log: log,
			Micro: stmf401re.Stmf401re{
				Log:                 log,
				LocalUUID:           localUUID,
				Firmware:            stmf401re.FIRMWARE{},
			},
		}, nil
	}
	return nil, fmt.Errorf("Unsupported board type")
}
