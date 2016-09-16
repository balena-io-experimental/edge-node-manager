package proxyvisor

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/device"
	"github.com/parnurzeal/gorequest"
)

var (
	address string
	version string
	key     string
)

// TODO handle no connection

// DependantApplicationsList returns all dependant applications assigned to the edge-node-manager
func DependantApplicationsList() (map[string]application.Application, []error) {
	url, err := buildPath(address, []string{version, "applications"})
	if err != nil {
		return nil, []error{err}
	}

	request := gorequest.New()
	request.Get(url)
	request.Query(key)
	resp, body, errs := request.EndBytes()

	if errs := handleResp(resp, errs); errs != nil {
		return nil, errs
	}

	var buffer []application.Application
	if err := json.Unmarshal(body, &buffer); err != nil {
		return nil, []error{err}
	}

	applications := make(map[string]application.Application)
	for _, application := range buffer {
		applications[application.Name] = application
	}

	return applications, nil
}

// DependantApplicationUpdate downloads the binary.tar for a specific application and commit,
// saving it to {ENM_ASSETS_DIRECTORY}/{appUUID}/{commit}/binary.tar
// Not convinced this works, perhaps try https://github.com/cavaliercoder/grab
func DependantApplicationUpdate(appUUID int, commit string) []error {
	url, err := buildPath(address, []string{version, "assets", strconv.Itoa(appUUID), commit})
	if err != nil {
		return []error{err}
	}

	request := gorequest.New()
	request.Get(url)
	request.Query(key)
	resp, body, errs := request.EndBytes()

	if errs := handleResp(resp, errs); errs != nil {
		return errs
	}

	filePath := config.GetAssetsDir()
	filePath = path.Join(filePath, strconv.Itoa(appUUID))
	filePath = path.Join(filePath, commit)
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return []error{err}
	}
	filePath = path.Join(filePath, "binary.tar")

	out, err := os.Create(filePath)
	if err != nil {
		return []error{err}
	}
	defer out.Close()

	if _, err = io.Copy(out, bytes.NewReader(body)); err != nil {
		return []error{err}
	}

	return nil
}

// DependantDeviceLog transmits a log message and timestamp for a specific device
func DependantDeviceLog(resinUUID, message string) []error {
	url, err := buildPath(address, []string{version, "devices", resinUUID, "logs"})
	if err != nil {
		return []error{err}
	}

	type dependantDeviceLog struct {
		Message   string `json:"message"`
		TimeStamp int64  `json:"timestamp"`
	}

	content := &dependantDeviceLog{
		Message:   message,
		TimeStamp: time.Now().UTC().Unix(),
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		return []error{err}
	}

	request := gorequest.New()
	request.Put(url)
	request.Query(key)
	request.Send((string)(bytes))
	resp, _, errs := request.End()

	return handleResp(resp, errs)
}

func DependantDeviceInfoUpdate(device device.Device) error {
	return nil
}

func DependantDeviceInfo(device *device.Device) error {
	return nil
}

func DependantDeviceProvision(appUUID int) (*device.Device, error) {
	return nil, nil
}

func DependantDevicesList() (map[string]*device.Device, error) {
	return nil, nil
}

func init() {
	address = config.GetSuperAddr()
	version = config.GetSuperAPIVer()
	key = config.GetSuperAPIKey()

	type apiKey struct {
		APIKey string `json:"apikey"`
	}

	content := &apiKey{
		APIKey: key,
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		log.WithFields(log.Fields{
			"Key":   key,
			"Error": err,
		}).Fatal("Unable to marshall API key")
	}
	key = (string)(bytes)
}

func buildPath(base string, paths []string) (string, error) {
	url, err := url.ParseRequestURI(address)
	if err != nil {
		return "", err
	}

	for _, p := range paths {
		url.Path = path.Join(url.Path, p)
	}

	return url.String(), nil
}

func handleResp(resp gorequest.Response, errs []error) []error {
	if errs != nil {
		return errs
	}

	if resp.StatusCode != 200 {
		return []error{fmt.Errorf("invalid response received: %s", resp.Status)}
	}

	log.WithFields(log.Fields{
		"Response": resp.Status,
	}).Debug("Valid response received")

	return nil
}
