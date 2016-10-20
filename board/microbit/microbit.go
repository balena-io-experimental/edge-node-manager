package microbit

import (
	"fmt"

	"github.com/josephroberts/edge-node-manager/micro/nrf51822"
)

type Microbit struct {
	Micro nrf51822.Nrf51822
}

func (b Microbit) Update(path string) error {
	fmt.Println("MicroBit update")
	return nil
}

func (b Microbit) Scan() (map[string]bool, error) {
	fmt.Println("MicroBit scan")
	return nil, nil
}

func (b Microbit) Online() (bool, error) {
	fmt.Println("MicroBit online")
	return false, nil
}

func (b Microbit) Restart() error {
	fmt.Println("MicroBit restart")
	return nil
}

func (b Microbit) Identify() error {
	fmt.Println("MicroBit identify")
	return nil
}
