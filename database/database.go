package database

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"log"
)

type Interface interface {
	Start() error
	Stop()
	Query(key, value string) ([]byte, error)
	Insert(device []byte) ([]byte, error)
	Update(key string, device []byte) error
}

type Database struct {
	Tiedot    string
	Directory string
	Port      string
}

func (d Database) Start() error {
	/*
		ENM uses the tiedot database
		https://github.com/HouzuoGuo/tiedot
	*/
	err := exec.Command(d.Tiedot, "-mode=httpd", "-dir="+d.Directory, "-port="+d.Port).Start()
	if err != nil {
		log.Println("Failed to start database")
		return err
	}

	// Allow time for the database to start up before initialising
	time.Sleep(1 * time.Second)

	return d.initialise()
}

func (d Database) initialise() error {
	resp, err := http.Get("http://localhost:" + d.Port + "/create?col=Devices")
	if err != nil {
		log.Println("Failed to create Devices collection")
		return err
	}
	resp.Body.Close()

	resp, err = http.Get("http://localhost:" + d.Port + "/index?col=Devices&path=applicationUUID")
	if err != nil {
		log.Println("Failed to add applicationUUID to index")
		return err
	}
	resp.Body.Close()

	resp, err = http.Get("http://localhost:" + d.Port + "/index?col=Devices&path=localUUID")
	if err != nil {
		log.Println("Failed to add localUUID to index")
		return err
	}
	resp.Body.Close()

	return nil
}

func (d Database) Stop() {
	resp, _ := http.Get("http://localhost:" + d.Port + "/shutdown")
	resp.Body.Close()
}

func (d Database) Query(key, value string) ([]byte, error) {
	resp, err := http.PostForm("http://localhost:"+d.Port+"/query?",
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

func (d Database) Insert(device []byte) ([]byte, error) {
	resp, err := http.PostForm("http://localhost:"+d.Port+"/insert?",
		url.Values{"doc": {string(device)}, "col": {"Devices"}})
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to parse return from database")
		return nil, err
	}

	return b, nil
}

func (d Database) Update(key string, device []byte) error {
	resp, err := http.PostForm("http://localhost:"+d.Port+"/update?",
		url.Values{"doc": {string(device)}, "id": {key}, "col": {"Devices"}})
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to parse return from database")
		return err
	}

	return err
}
