package main

import (
	"net/http"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/api"
	"github.com/resin-io/edge-node-manager/application"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/process"
	"github.com/resin-io/edge-node-manager/supervisor"
)

func main() {
	log.Info("Starting edge-node-manager")

	supervisor.WaitUntilReady()

	delay, err := config.GetLoopDelay()
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load loop delay")
	}

	for {
		application.Load()

		// Sort applications to ensure they run in order
		var keys []int
		for key := range application.List {
			keys = append(keys, key)
		}
		sort.Ints(keys)

		for _, key := range keys {
			application := application.List[key]
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
	log.SetFormatter(&log.TextFormatter{ForceColors: true, DisableTimestamp: true})

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
