package proxyvisor

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
)

var (
	address string
	key     string
)

func DependantApplicationsList() (map[string]*application.Application, error) {
	target, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, err
	}

	target.Path = "/v1/applications/"
	data := url.Values{}
	data.Set("apikey", key)

	client := &http.Client{}
	req, err := http.NewRequest("GET", target.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	buffer := make([]application.Application, 0, 10)
	if err := json.Unmarshal(body, &buffer); err != nil {
		return nil, err
	}

	applications := make(map[string]*application.Application)
	for _, application := range buffer {
		applications[application.Name] = &application
	}

	return applications, nil
}

// NewDevice creates a new device on the resin dashboard and returns its UUID
func NewDevice(appUUID int) (string, error) {
	// Simulate proxyvisor whilst we wait for it to be released by returning a random 62 char string
	return random(62), nil
}

func random(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func init() {
	address = config.GetResinSupervisorAddress()
	key = config.GetResinSupervisorAPIKey()
}
