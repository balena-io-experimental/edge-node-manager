package proxyvisor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/josephroberts/edge-node-manager/application"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/josephroberts/edge-node-manager/device"
)

var (
	address string
	key     string
	version string
)

// TODO handle no connection

// DependantApplicationsList returns all dependant applications assigned to the edge-node-manager
func DependantApplicationsList() (map[string]application.Application, error) { //TODO set req type
	target, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, err
	}

	target.Path = buildPath([]string{version, "applications"})
	data := url.Values{}
	data.Set("apikey", key)

	client := &http.Client{}
	req, err := http.NewRequest("GET", target.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	logReq(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid response received: %s", resp.Status)
	}
	log.WithFields(log.Fields{
		"Response": resp.Status,
	}).Debug("Valid response received")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var buffer []application.Application
	if err := json.Unmarshal(body, &buffer); err != nil {
		return nil, err
	}

	applications := make(map[string]application.Application)
	for _, application := range buffer {
		applications[application.Name] = application
	}

	return applications, nil
}

// DependantApplicationUpdate downloads the binary.tar for a specific application and commit,
// saving it to {ENM_ASSETS_DIRECTORY}/{appUUID}/{commit}/binary.tar
func DependantApplicationUpdate(appUUID int, commit string) error {
	target, err := url.ParseRequestURI(address)
	if err != nil {
		return err
	}

	target.Path = buildPath([]string{version, "assets", strconv.Itoa(appUUID), commit})
	data := url.Values{}
	data.Set("apikey", key)

	client := &http.Client{}
	req, err := http.NewRequest("GET", target.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	logReq(req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid response received: %s", resp.Status)
	}
	log.WithFields(log.Fields{
		"Response": resp.Status,
	}).Debug("Valid response received")

	filePath := config.GetAssetsDir()
	filePath = path.Join(filePath, strconv.Itoa(appUUID))
	filePath = path.Join(filePath, commit)
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return err
	}

	// TODO delete old commit after download

	filePath = path.Join(filePath, "binary.tar")
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

func DependantDeviceLog(resinUUID, message string) error {
	target, err := url.ParseRequestURI(address)
	if err != nil {
		return err
	}
	target.Path = buildPath([]string{version, "devices", resinUUID, "logs"})

	data := url.Values{}
	data.Set("apikey", key)
	//data.Set("message", message)
	//data.Set("timestamp", strconv.FormatInt(time.Now().UTC().Unix(), 10))

	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("PUT", target.String(), bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	logReq(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid response received: %s", resp.Status)
	}
	log.WithFields(log.Fields{
		"Response": resp.Status,
	}).Debug("Valid response received")

	return nil
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
	key = config.GetSuperAPIKey()
	version = config.GetSuperAPIVer()
}

func buildPath(paths []string) string {
	var u url.URL
	for _, p := range paths {
		u.Path = path.Join(u.Path, p)
	}
	return u.String()
}

func logReq(req *http.Request) error {
	if log.GetLevel() != log.DebugLevel {
		return nil
	}

	requestDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Request dump": (string)(requestDump),
	}).Debug("HTTP request")

	return nil
}
