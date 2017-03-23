package application

import (
	"encoding/json"
	"fmt"

	"github.com/resin-io/edge-node-manager/board"
	"github.com/resin-io/edge-node-manager/board/microbit"
	"github.com/resin-io/edge-node-manager/board/nrf51822dk"
)

type Application struct {
	Board        board.Interface        `json:"-"`
	BoardType    board.Type             `json:"-"`
	Name         string                 `json:"name"`
	ResinUUID    int                    `json:"id"`
	Commit       string                 `json:"-"`      // Ignore this when unmarshalling from the supervisor as we want to set the target commit
	TargetCommit string                 `json:"commit"` // Set json tag to commit as the supervisor has no concept of target commit
	Config       map[string]interface{} `json:"config"`
}

func (a Application) String() string {
	return fmt.Sprintf(
		"Board type: %s, "+
			"Name: %s, "+
			"Resin UUID: %d, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Config: %v",
		a.BoardType,
		a.Name,
		a.ResinUUID,
		a.Commit,
		a.TargetCommit,
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
