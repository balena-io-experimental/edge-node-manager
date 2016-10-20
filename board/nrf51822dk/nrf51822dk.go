package nrf51822dk

import (
	"fmt"

	"github.com/josephroberts/edge-node-manager/micro/nrf51822"
)

type Nrf51822dk struct {
	Micro nrf51822.Nrf51822
}

func (b Nrf51822dk) Update(path string) error {
	fmt.Println("nRF51822-DK update")
	return nil
}

func (b Nrf51822dk) Scan() (map[string]bool, error) {
	fmt.Println("nRF51822-DK scan")
	return nil, nil
}

func (b Nrf51822dk) Online() (bool, error) {
	fmt.Println("nRF51822-DK online")
	return false, nil
}

func (b Nrf51822dk) Restart() error {
	fmt.Println("nRF51822-DK restart")
	return nil
}

func (b Nrf51822dk) Identify() error {
	fmt.Println("nRF51822-DK identify")
	return nil
}
