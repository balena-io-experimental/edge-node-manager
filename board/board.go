package board

import (
	"github.com/josephroberts/edge-node-manager/board/microbit"
	"github.com/josephroberts/edge-node-manager/micro/nrf51822"
)

type Type string

const (
	MICROBIT Type = "MicroBit"
)

type Interface interface {
	Update(path string) error
	Scan() (map[string]bool, error)
	Online() (bool, error)
	Restart() error
	Identify() error
}

func Create(boardType Type) Interface {
	switch boardType {
	case MICROBIT:
		return microbit.MicroBit{
			Micro: nrf51822.Nrf51822{},
		}
	}
	return nil
}
