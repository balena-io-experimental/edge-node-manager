package application

import (
	"fmt"

	"github.com/josephroberts/edge-node-manager/device"
	"github.com/josephroberts/edge-node-manager/micro"
	"github.com/josephroberts/edge-node-manager/proxyvisor"
	"github.com/josephroberts/edge-node-manager/radio"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
)

var (
	// List holds all the applications assigned to the edge-node-manager
	List map[int]*Application
)

// Application contains all the variables needed to define an application
type Application struct {
	UUID         int         `json:"appId"`
	Name         string      `json:"name"`
	Commit       string      `json:"-"`
	TargetCommit string      `json:"commit"`
	Env          interface{} `json:"env"`
	DeviceType   string      `json:"device_type"`
	device.Type  `json:"type"`
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
	List = make(map[int]*Application)

	bytes, errs := proxyvisor.DependantApplicationsList()
	if errs != nil {
		log.WithFields(log.Fields{
			"Errors": errs,
		}).Fatal("Unable to get the dependant application list")
	}

	var buffer []Application
	if err := json.Unmarshal(bytes, &buffer); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to unmarshal the dependant application list")
	}

	for key := range buffer {
		UUID := buffer[key].UUID
		List[UUID] = &buffer[key]
	}

	initApp(13015, micro.NRF51822, radio.BLUETOOTH)
}

func initApp(UUID int, micro micro.Type, radio radio.Type) {
	if _, exists := List[UUID]; !exists {
		log.WithFields(log.Fields{
			"UUID": UUID,
		}).Fatal("Application does not exist")
	}

	List[UUID].Type = device.Type{
		Micro: micro,
		Radio: radio,
	}
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
