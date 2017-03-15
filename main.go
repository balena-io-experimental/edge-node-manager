package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/jsonq"
	"github.com/resin-io/edge-node-manager/api"
	"github.com/resin-io/edge-node-manager/application"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/process"
	"github.com/resin-io/edge-node-manager/supervisor"
)

// This variable will be populated at build time with the current version tag
var version string

func main() {
	if err := checkVersion(); err != nil {
		log.Error("Unable to check if edge-node-manager is up to date")
	}

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

func checkVersion() error {
	resp, err := http.Get("https://api.github.com/repos/resin-io/edge-node-manager/releases/latest")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	latest, err := jsonq.NewQuery(data).String("tag_name")
	if err != nil {
		return err
	}

	if version == latest {
		return nil
	}

	log.WithFields(log.Fields{
		"Current version": version,
		"Latest version":  latest,
		"Update command":  "git push resin master:resin-nocache",
	}).Warn("Please update edge-node-manager")

	return nil
}
