package application

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/device"
	tarinator "github.com/verybluebot/tarinator-go"
)

// Application contains all the variables needed to run the application
type Application struct {
	UUID       int         `json:"id"`
	Name       string      `json:"name"`
	Commit     string      `json:"commit"`
	Env        interface{} `json:"env"`
	DeviceType string      `json:"device_type"` // not used as always set to edge - see device.Type instead
	device.Type
	Directory string
}

func (a Application) String() string {
	return fmt.Sprintf(
		"UUID: %d, "+
			"Name: %s, "+
			"Commit: %s, "+
			"Env: %v, "+
			"Device Type: %s, "+
			"Micro Type: %s, "+
			"Radio Type: %s, "+
			"Directory: %s, ",
		a.UUID,
		a.Name,
		a.Commit,
		a.Env,
		a.DeviceType,
		a.Type.Micro,
		a.Type.Radio,
		a.Directory)
}

// ParseCommit finds and extracts the firmware tar belonging to this application
func (a *Application) ParseCommit() error {
	appDir := filepath.Join(config.GetPersistantDirectory(), (string)(a.UUID))

	commitDirectories, err := ioutil.ReadDir(appDir)
	if err != nil {
		return err
	} else if len(commitDirectories) == 0 {
		return nil
	} else if len(commitDirectories) > 1 {
		return errors.New("More than one commit found")
	}

	commit := commitDirectories[0].Name()
	if a.Commit == commit {
		return nil
	}

	a.Commit = commit
	a.Directory = path.Join(appDir, a.Commit)

	tarPath := filepath.Join(a.Directory, "binary.tar")
	return tarinator.UnTarinate(a.Directory, tarPath)
}
