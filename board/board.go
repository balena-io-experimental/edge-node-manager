package board

import (
	"github.com/josephroberts/edge-node-manager/board/microbit"
	"github.com/josephroberts/edge-node-manager/board/nrf51822dk"
	"github.com/josephroberts/edge-node-manager/micro/nrf51822"
)

type Type string

const (
	MICROBIT   Type = "MicroBit"
	NRF51822DK      = "nRF51822-DK"
)

type Interface interface {
	Update(path string) error
	Scan(applicationUUID int) (map[string]bool, error)
	Online() (bool, error)
	Restart() error
	Identify() error
}

func Create(boardType Type, localUUID string) Interface {
	switch boardType {
	case MICROBIT:
		return microbit.Microbit{
			Micro: nrf51822.Nrf51822{
				LocalUUID: localUUID,
			},
		}
	case NRF51822DK:
		return nrf51822dk.Nrf51822dk{
			Micro: nrf51822.Nrf51822{
				LocalUUID: localUUID,
			},
		}
	}
	return nil
}
