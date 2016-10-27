package main

import (
	"github.com/resin-io/edge-node-manager/board"
	"github.com/resin-io/edge-node-manager/device"
)

func main() {
	d, _ := device.Create(board.MICROBIT, "test", "test", "f054e63ee37b35b4585916292566cb3f0c073ee4219e452ba65adbaf363509", 1, "test", "", nil, nil)
	d.Board.Update("")
}

// 	delay, err := config.GetLoopDelay()
// 	if err != nil {
// 		log.WithFields(log.Fields{
// 			"Error": err,
// 		}).Fatal("Unable to load loop delay")
// 	}

// 	for {
// 		application.Load()

// 		for _, a := range application.List {
// 			if errs := process.Run(a); errs != nil {
// 				log.WithFields(log.Fields{
// 					"Application": a,
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
