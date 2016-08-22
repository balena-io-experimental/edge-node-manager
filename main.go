package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/radio"
)

func main() {
	db := &database.Database{
		Tiedot:    config.GetTiedot(),
		Directory: path.Join(config.GetPersistantDirectory(), config.GetDbDirectory()),
		Port:      config.GetDbPort(),
	}
	tidyOnExit(db)
	db.Start()

	radios := make(map[radio.SupportedRadio]radio.Interface)
	radios[radio.BLUETOOTH] = radio.Create(radio.BLUETOOTH)
	radios[radio.WIFI] = radio.Create(radio.WIFI)
	radios[radio.ZIGBEE] = radio.Create(radio.ZIGBEE)

	applications := make(map[string]application.Interface)
	applications["resin"] = &application.Application{
		Name:      "resin",
		Directory: config.GetPersistantDirectory(),
		Database:  db,
		Device:    device.NRF51822,
		Radio:     radios[radio.BLUETOOTH],
	}

	loopDelay, err := config.GetLoopDelay()
	if err != nil {
		log.Fatal("Failed to get the loop delay")
	}

	for {
		for _, application := range applications {
			application.Process()
		}

		// Delay between processing each set of applications to prevent 100% CPU usage
		time.Sleep(loopDelay * time.Second)
	}
}

func tidyOnExit(db *database.Database) {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	signal.Notify(channel, syscall.SIGTERM)
	go func() {
		<-channel
		db.Stop()
		os.Exit(0)
	}()
}
