package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/jmoiron/jsonq"
	"github.com/resin-io/edge-node-manager/api"
	"github.com/resin-io/edge-node-manager/application"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/device"
	"github.com/resin-io/edge-node-manager/process"
	"github.com/resin-io/edge-node-manager/radio/bluetooth"
	"github.com/resin-io/edge-node-manager/supervisor"
)

var (
	// This variable will be populated at build time with the current version tag
	version string
	// This variable defines the delay between each processing loop
	loopDelay time.Duration
)

func main() {
	log.Info("Starting edge-node-manager")

	if err := checkVersion(); err != nil {
		log.Error("Unable to check if edge-node-manager is up to date")
	}

	supervisor.WaitUntilReady()

	for {
		// Run processing loop
		loop()

		// Delay between processing each set of applications to prevent 100% CPU usage
		time.Sleep(loopDelay)
	}
}

func init() {
	log.SetLevel(config.GetLogLevel())
	log.SetFormatter(&log.TextFormatter{ForceColors: true, DisableTimestamp: true})

	var err error
	loopDelay, err = config.GetLoopDelay()
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load loop delay")
	}

	dbDir := config.GetDbDir()
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"Directory": dbDir,
			"Error":     err,
		}).Fatal("Unable to create database directory")
	}

	db, err := storm.Open(config.GetDbPath())
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to open database")
	}
	defer db.Close()

	if err := db.Init(&device.Device{}); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to initialise database")
	}

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

func loop() {
	// Get applications from the supervisor
	bytes, errs := supervisor.DependentApplicationsList()
	if errs != nil {
		log.WithFields(log.Fields{
			"Errors": errs,
		}).Error("Unable to get applications")
		return
	}

	// Unmarshal applications
	applications, err := application.Unmarshal(bytes)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to unmarshal applications")
		return
	}

	// Sort applications to ensure they run in order
	var keys []int
	for key := range applications {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	// Process applications
	for _, key := range keys {
		if errs := process.Run(applications[key]); errs != nil {
			log.WithFields(log.Fields{
				"Application": applications[key],
				"Errors":      errs,
			}).Error("Unable to process application")
		}
	}

	// Reset the bluetooth device to clean up any left over go routines etc. Quick fix
	if err := bluetooth.ResetDevice(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to reset bluetooth")
		return
	}
}
