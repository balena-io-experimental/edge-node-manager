package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/radio"
)

func main() {
	tidyOnExit()

	database.Start()

	radios := make(map[radio.SupportedRadio]radio.RadioInterface)
	radios[radio.BLUETOOTH] = &radio.Bluetooth{}
	radios[radio.WIFI] = &radio.Wifi{}
	radios[radio.ZIGBEE] = &radio.Zigbee{}

	applications := make(map[string]application.ApplicationInterface)
	applications["applicationA"] = &application.Application{"applicationA", device.NRF51822, radios[radio.BLUETOOTH]}
	applications["applicationB"] = &application.Application{"applicationB", device.ESP8266, radios[radio.WIFI]}
	applications["applicationC"] = &application.Application{"applicationC", device.MICROBIT, radios[radio.BLUETOOTH]}

	for {
		for _, application := range applications {
			application.Process()
		}

		time.Sleep(5 * time.Second)
	}
}

func tidyOnExit() {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	signal.Notify(channel, syscall.SIGTERM)
	go func() {
		<-channel
		database.Stop()
		os.Exit(0)
	}()
}
