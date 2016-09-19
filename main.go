package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/api"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/database"
)

func main() {
	log.Info("Starting edge node manager")

	// apps, errs := proxyvisor.DependantApplicationsList()
	// if errs != nil {
	// 	log.WithFields(log.Fields{
	// 		"Errors": errs,
	// 	}).Fatal("Unable to get the dependant application list")
	// }
	// for key, app := range apps {
	// 	log.WithFields(log.Fields{
	// 		"Key":   key,
	// 		"Value": app,
	// 	}).Debug("Application")
	// }

	// err := proxyvisor.DependantApplicationUpdate(13015, "d43bea5e16658e653088ce4b9a91b6606c3c2a0d")
	// if err != nil {
	// 	log.WithFields(log.Fields{
	// 		"Error": err,
	// 	}).Fatal("Unable to get the dependant application update")
	// }

	// errs := proxyvisor.DependantDeviceLog("fef6e0b23f65ecef1c10bd49ef155694720194940f3e990477f7b21d54ddfa", "hello")
	// if errs != nil {
	// 	log.WithFields(log.Fields{
	// 		"Errors": errs,
	// 	}).Fatal("Unable to set the dependant device log")
	// }

	// errs := proxyvisor.DependantDeviceInfoUpdate("fef6e0b23f65ecef1c10bd49ef155694720194940f3e990477f7b21d54ddfa", "updating", true)
	// if errs != nil {
	// 	log.WithFields(log.Fields{
	// 		"Errors": errs,
	// 	}).Fatal("Unable to set the dependant device info")
	// }

	router := api.NewRouter()
	log.Fatal(http.ListenAndServe(config.GetENMAddr(), router))

	// nrf51822 := device.Type{
	// 	Micro: micro.NRF51822,
	// 	Radio: radio.BLUETOOTH,
	// }
	// esp8266 := device.Type{
	// 	Micro: micro.ESP8266,
	// 	Radio: radio.WIFI,
	// }

	// apps := []*application.Application{
	// 	&application.Application{
	// 		Name: "resin",
	// 		Type: nrf51822,
	// 	},
	// 	&application.Application{
	// 		Name: "resin_esp8266",
	// 		Type: esp8266,
	// 	}}

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
	// 		if err := app.Process(); err != nil {
	// 			log.WithFields(log.Fields{
	// 				"Application UUID": app.UUID,
	// 				"Error":            err,
	// 			}).Fatal("Unable to process application")
	// 		}
	// 	}

	// 	// Delay between processing each set of applications to prevent 100% CPU usage
	// 	time.Sleep(delay * time.Second)
	// }
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)

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
