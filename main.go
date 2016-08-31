package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/radio"
)

func main() {
	log.Info("Starting edge node manager")

	nrf51822 := device.DeviceType{
		Micro: micro.NRF51822,
		Radio: radio.BLUETOOTH,
	}
	esp8266 := device.DeviceType{
		Micro: micro.ESP8266,
		Radio: radio.WIFI,
	}

	apps := []*application.Application{
		&application.Application{
			UUID:       "resin",
			DeviceType: nrf51822,
		},
		&application.Application{
			UUID:       "resin_esp8266",
			DeviceType: esp8266,
		}}

	delay, err := config.GetLoopDelay()
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load loop delay")
	}

	log.WithFields(log.Fields{
		"Loop delay": delay,
	}).Info("Started edge node manager")

	for {
		for _, app := range apps {
			if err := app.Process(); err != nil {
				log.WithFields(log.Fields{
					"Application UUID": app.UUID,
					"Error":            err,
				}).Fatal("Unable to process application")
			}
		}

		// Delay between processing each set of applications to prevent 100% CPU usage
		time.Sleep(delay * time.Second)
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.InfoLevel)

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	signal.Notify(channel, syscall.SIGTERM)
	go func() {
		<-channel
		if err := database.Stop(); err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("Unable to stop database")
		}

		// if err := bluetooth.Disconnect; err != nil {
		// 	log.WithFields(log.Fields{
		// 		"Error": err,
		// 	}).Fatal("Unable to disconnect bluetooth")
		// }

		os.Exit(0)
	}()
}
