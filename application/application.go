package application

import (
	"fmt"

	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/proxyvisor"
	"github.com/josephroberts/edge-node-manager/radio"
	"github.com/mitchellh/mapstructure"

	log "github.com/Sirupsen/logrus"
)

// List holds all the applications assigned to the edge-node-manager
var List map[string]*Application

// Application contains all the variables needed to define an application
type Application struct {
	UUID         int         `mapstructure:"appId" structs:"appId"`
	Name         string      `mapstructure:"name" structs:"name"`
	Commit       string      `mapstructure:"commit" structs:"commit"`
	TargetCommit string      `mapstructure:"targetCommit" structs:"targetCommit"`
	Env          interface{} `mapstructure:"env" structs:"env"`
	DeviceType   string      `mapstructure:"device_type" structs:"device_type"`
	device.Type  `mapstructure:"type" structs:"type"`
	// Directory string
}

func (a Application) String() string {
	return fmt.Sprintf(
		"UUID: %d, "+
			"Name: %s, "+
			"Commit: %s, "+
			"Target commit: %s, "+
			"Env: %v, "+
			"Device type: %s, "+
			"Micro type: %s, "+
			"Radio type: %s", //+
		// "Directory: %s, ",
		a.UUID,
		a.Name,
		a.Commit,
		a.TargetCommit,
		a.Env,
		a.DeviceType,
		a.Type.Micro,
		a.Type.Radio) //,
	// a.Directory)
}

func init() {
	buffer, errs := proxyvisor.DependantApplicationsList()
	if errs != nil {
		log.WithFields(log.Fields{
			"Errors": errs,
		}).Fatal("Unable to get the dependant application list")
	}

	List = make(map[string]*Application)

	for _, item := range buffer {
		var app Application
		if err := mapstructure.Decode(item, &app); err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("Unable to decode the dependant application list")
		}
		List[app.Name] = &app
	}

	if _, exists := List["resin"]; !exists {
		log.WithFields(log.Fields{
			"Key": "resin",
		}).Fatal("Application does not exist")
	}

	nrf51822 := device.Type{
		Micro: micro.NRF51822,
		Radio: radio.BLUETOOTH,
	}
	List["resin"].Type = nrf51822
}

// ParseCommit finds and extracts the firmware tar belonging to this application
// func (a *Application) ParseCommit() error {
// 	filePath := config.GetAssetsDir()
// 	filePath = path.Join(filePath, strconv.Itoa(a.UUID))

// 	commitDirectories, err := ioutil.ReadDir(filePath)
// 	if err != nil {
// 		return err
// 	} else if len(commitDirectories) == 0 {
// 		return nil
// 	} else if len(commitDirectories) > 1 {
// 		return errors.New("More than one commit found")
// 	}

// 	commit := commitDirectories[0].Name()
// 	if a.Commit == commit {
// 		return nil
// 	}

// 	a.Commit = commit
// 	a.Directory = path.Join(filePath, a.Commit)

// 	tarPath := filepath.Join(a.Directory, "binary.tar")
// 	return tarinator.UnTarinate(a.Directory, tarPath)
// }
