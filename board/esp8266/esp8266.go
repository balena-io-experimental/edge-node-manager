package esp8266

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/radio/wifi"
)

type Esp8266 struct {
	Log       *log.Logger
	LocalUUID string
}

func (b Esp8266) InitialiseRadio() error {
	return wifi.Initialise()
}

func (b Esp8266) CleanupRadio() error {
	return wifi.Cleanup()
}

func (b Esp8266) Update(filePath string) error {
	b.Log.Info("Starting update")

	ip, err := wifi.GetIP(b.LocalUUID)
	if err != nil {
		return err
	}

	if err := wifi.PostForm("http://"+ip+"/update", path.Join(filePath, "firmware.bin")); err != nil {
		return err
	}

	b.Log.Info("Finished update")

	return nil
}

func (b Esp8266) Scan(applicationUUID int) (map[string]struct{}, error) {
	return wifi.Scan(strconv.Itoa(applicationUUID))
}

func (b Esp8266) Online() (bool, error) {
	return wifi.Online(b.LocalUUID)
}

func (b Esp8266) Restart() error {
	b.Log.Info("Restarting...")
	return fmt.Errorf("Restart not implemented")
}

func (b Esp8266) Identify() error {
	b.Log.Info("Identifying...")
	return fmt.Errorf("Identify not implemented")
}

func (b Esp8266) UpdateEnvironment(env interface{}) error {
	b.Log.WithFields(log.Fields{
		"Environment": env,
	}).Info("Updating environment...")

	buffer, err := json.Marshal(env)
	if err != nil {
		return err
	}

	// post Request
	fmt.Println(string(buffer))


	return fmt.Errorf("woooooooo")
}
