package database

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/josephroberts/edge-node-manager/config"

	"log"
)

func Start() error {
	err := exec.Command(os.Getenv("GOPATH")+"/bin/tiedot", "-mode=httpd", "-dir="+filepath.Join(config.Persistant, config.Db), "-port="+config.DbPort).Start()
	if err != nil {
		log.Println("Failed to start database")
		return err
	}

	time.Sleep(1 * time.Second)

	return initialise()
}

func initialise() error {
	resp, err := http.Get("http://localhost:" + config.DbPort + "/create?col=Devices")
	if err != nil {
		log.Println("Failed to create Devices collection")
		return err
	}
	resp.Body.Close()

	resp, err = http.Get("http://localhost:" + config.DbPort + "/index?col=Devices&path=applicationUUID")
	if err != nil {
		log.Println("Failed to add applicationUUID to index")
		return err
	}
	resp.Body.Close()

	resp, err = http.Get("http://localhost:" + config.DbPort + "/index?col=Devices&path=localUUID")
	if err != nil {
		log.Println("Failed to add localUUID to index")
		return err
	}
	resp.Body.Close()

	return nil
}

func Stop() {
	resp, _ := http.Get("http://localhost:" + config.DbPort + "/shutdown")
	resp.Body.Close()
}

func Query(key, value string) ([]byte, error) {
	resp, err := http.PostForm("http://localhost:"+config.DbPort+"/query?col=Devices",
		url.Values{"q": {`{"eq": "` + value + `", "in": ["` + key + `"]}`}, "col": {"Devices"}})
	defer resp.Body.Close()

	if err != nil {
		log.Println("Failed to query database")
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to parse return from database")
		return nil, err
	}

	return b, nil
}

func Insert(device []byte) ([]byte, error) {
	resp, err := http.PostForm("http://localhost:"+config.DbPort+"/insert?col=Devices",
		url.Values{"doc": {string(device)}, "col": {"Devices"}})
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to parse return from database")
		return nil, err
	}

	return b, nil
}
