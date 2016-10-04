package main

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/josephroberts/edge-node-manager/api"
	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/process"
)

// Uses the logrus package
// https://github.com/Sirupsen/logrus

func main() {
	log.Info("Starting Edge-node-manager")

	for _, value := range application.List {
		log.WithFields(log.Fields{
			"Application": value,
		}).Info("Edge-node-manager application")
	}

	delay, err := config.GetLoopDelay()
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load loop delay")
	}

	log.WithFields(log.Fields{
		"Loop delay": delay,
	}).Info("Started Edge-node-manager")

	for {
		for _, application := range application.List {
			if errs := process.Run(application); errs != nil {
				log.WithFields(log.Fields{
					"Application": application,
					"Errors":      errs,
				}).Error("Unable to process application")
			}
		}

		// Delay between processing each set of applications to prevent 100% CPU usage
		time.Sleep(delay * time.Second)
	}
}

func init() {
	log.SetLevel(config.GetLogLevel())

	go func() {
		router := api.NewRouter()

		port, err := config.GetENMPort()
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("Unable to get ENM port")
		}

		log.WithFields(log.Fields{
			"Port": port,
		}).Debug("Initialising incoming supervisor API")

		if err := http.ListenAndServe(port, router); err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("Unable to initialise incoming supervisor API")
		}
	}()
}
