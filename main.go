package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/api"
	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
)

func main() {
	log.Info("Starting edge node manager")

	fmt.Println(application.List)

	// Get app list
	// Set device type - hardcoded
	// Write into DB
	// Before each loop read DB
	// Set app target from handler
	// In app check if target matches commit, extract if not

	// apps, errs := proxyvisor.DependantApplicationsList()
	// if errs != nil {
	// 	log.WithFields(log.Fields{
	// 		"Errors": errs,
	// 	}).Fatal("Unable to get the dependant application list")
	// }

	// if _, exists := apps["resin"]; exists != true {
	// 	log.WithFields(log.Fields{
	// 		"Key": "resin",
	// 	}).Fatal("Application does not exist")
	// }

	// nrf51822 := device.Type{
	// 	Micro: micro.NRF51822,
	// 	Radio: radio.BLUETOOTH,
	// }
	// apps["resin"].Type = nrf51822

	// if log.GetLevel() == log.DebugLevel {
	// 	for key, app := range apps {
	// 		log.WithFields(log.Fields{
	// 			"Key":   key,
	// 			"Value": app,
	// 		}).Debug("Dependant applications")
	// 	}
	// }

	// delay, err := config.GetLoopDelay()
	// if err != nil {
	// 	log.WithFields(log.Fields{
	// 		"Error": err,
	// 	}).Fatal("Unable to load loop delay")
	// }

	// log.WithFields(log.Fields{
	// 	"Loop delay": delay,
	// }).Info("Started edge node manager")

	// for {
	// 	for _, app := range apps {
	// 		if errs := process.Run(app); errs != nil {
	// 			log.WithFields(log.Fields{
	// 				"Application": app,
	// 				"Errors":      errs,
	// 			}).Error("Unable to process application")
	// 		}
	// 	}

	// 	// Delay between processing each set of applications to prevent 100% CPU usage
	// 	time.Sleep(delay * time.Second)
	// }
}

func init() {
	//log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)

	go func() {
		router := api.NewRouter()
		if err := http.ListenAndServe(config.GetENMAddr(), router); err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("Unable to start API server")
		}
	}()

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

		os.Exit(0)
	}()
}
