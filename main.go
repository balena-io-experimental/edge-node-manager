package main

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/api"
	"github.com/resin-io/edge-node-manager/application"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/process"
)

func main() {
	delay, err := config.GetLoopDelay()
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load loop delay")
	}

	for {
		application.Load()

		for _, a := range application.List {
			if errs := process.Run(a); errs != nil {
				log.WithFields(log.Fields{
					"Application": a,
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
		port := ":1337"

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
