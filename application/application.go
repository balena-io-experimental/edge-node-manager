package application

import (
	"encoding/json"
	"fmt"

	"github.com/resin-io/edge-node-manager/board"
	"github.com/resin-io/edge-node-manager/board/esp8266"
	"github.com/resin-io/edge-node-manager/board/microbit"
	"github.com/resin-io/edge-node-manager/board/nrf51822dk"
)

type Application struct {
	Board     board.Interface        `json:"-"`
	BoardType board.Type             `json:"-"`
	Name      string                 `json:"name"`
	ResinUUID int                    `json:"id"`
	Config    map[string]interface{} `json:"config"`
}

func (a Application) String() string {
	return fmt.Sprintf(
		"Board type: %s, "+
			"Name: %s, "+
			"Resin UUID: %d, "+
			"Config: %v",
		a.BoardType,
		a.Name,
		a.ResinUUID,
		a.Config)
}

func Unmarshal(bytes []byte) (map[int]Application, error) {
	applications := make(map[int]Application)

	var buffer []Application
	if err := json.Unmarshal(bytes, &buffer); err != nil {
		return nil, err
	}

	for key, value := range buffer {
		value, ok := value.Config["RESIN_HOST_TYPE"]
		if !ok {
			continue
		}
		boardType := (board.Type)(value.(string))

		var b board.Interface
		switch boardType {
		case board.MICROBIT:
			b = microbit.Microbit{}
		case board.NRF51822DK:
			b = nrf51822dk.Nrf51822dk{}
		case board.ESP8266:
			b = esp8266.Esp8266{}
		default:
			continue
		}

		application := buffer[key]
		application.BoardType = boardType
		application.Board = b

		applications[application.ResinUUID] = application
	}

	return applications, nil
}
