package main

import (
	"fmt"

	"github.com/josephroberts/edge-node-manager/board"
	"github.com/josephroberts/edge-node-manager/device"
)

func main() {
	d := device.Create(board.MICROBIT, "test_name", "test_localUUID", "test_resinUUID", 123456789, "test_applicationName", "test_targetCommit", nil, nil)
	fmt.Println(d)
	d.Board.Update("")

	d = device.Create(board.NRF51822DK, "test_name", "test_localUUID", "test_resinUUID", 123456789, "test_applicationName", "test_targetCommit", nil, nil)
	fmt.Println(d)
	d.Board.Update("")
}

// 	log.Info("Starting edge-node-manager")

// 	log.WithFields(log.Fields{
// 		"Number": len(application.List),
// 	}).Info("edge-node-manager applications")

// 	delay, err := config.GetLoopDelay()
// 	if err != nil {
// 		log.WithFields(log.Fields{
// 			"Error": err,
// 		}).Fatal("Unable to load loop delay")
// 	}

// 	log.Info("Started edge-node-manager")

// 	for {
// 		for _, application := range application.List {
// 			if errs := process.Run(application); errs != nil {
// 				log.WithFields(log.Fields{
// 					"Application": application,
// 					"Errors":      errs,
// 				}).Error("Unable to process application")
// 			}
// 		}

// 		// Delay between processing each set of applications to prevent 100% CPU usage
// 		time.Sleep(delay * time.Second)
// 	}
// }

// func init() {
// 	log.SetLevel(config.GetLogLevel())

// 	go func() {
// 		router := api.NewRouter()
// 		port := ":1337"

// 		log.WithFields(log.Fields{
// 			"Port": port,
// 		}).Debug("Initialising incoming supervisor API")

// 		if err := http.ListenAndServe(port, router); err != nil {
// 			log.WithFields(log.Fields{
// 				"Error": err,
// 			}).Fatal("Unable to initialise incoming supervisor API")
// 		}
// 	}()
// }
