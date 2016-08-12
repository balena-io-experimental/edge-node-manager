package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/device/deviceType"
	"github.com/josephroberts/edge-node-manager/radio/radioType"
)

func main() {
	exit()

	nrf51822 := &device.Type{
		Device: deviceType.NRF51822,
		Radio:  radioType.BLUETOOTH,
	}
	esp8266 := &device.Type{
		Device: deviceType.ESP8266,
		Radio:  radioType.WIFI,
	}

	applications := make([]*application.Application, 0, 10)
	applications = append(applications,
		&application.Application{
			UUID:       "resin",
			DeviceType: nrf51822,
		})
	applications = append(applications,
		&application.Application{
			UUID:       "resin_esp8266",
			DeviceType: esp8266,
		})

	if delay, err := config.GetLoopDelay(); err != nil {
		log.Fatalf("Unable to load loop delay: %v", err)
	} else {
		for {
			for _, application := range applications {
				if err := application.Process(); err != nil {
					log.Printf("Unable to process application %s: %v", application.UUID, err)
				}
			}

			// Delay between processing each set of applications to prevent 100% CPU usage
			time.Sleep(delay * time.Second)
		}
	}
}

func exit() {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	signal.Notify(channel, syscall.SIGTERM)
	go func() {
		<-channel
		database.Stop()
		os.Exit(0)
	}()
}
