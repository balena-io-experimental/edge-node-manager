package supervisor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cavaliercoder/grab"
	"github.com/josephroberts/edge-node-manager/config"
	"github.com/parnurzeal/gorequest"
)

// Uses the grab package
// https://github.com/cavaliercoder/grab
// Uses the gorequest package
// https://github.com/parnurzeal/gorequest

var (
	address string
	version string
	key     string
	rawKey  string
)

// DependantApplicationsList returns all dependant applications assigned to the edge-node-manager
func DependantApplicationsList() ([]byte, []error) {
	url, err := buildPath(address, []string{version, "dependent-apps"})
	if err != nil {
		return nil, []error{err}
	}

	req := gorequest.New()
	req.Get(url)
	req.Query(key)

	log.WithFields(log.Fields{
		"URL":    req.Url,
		"Method": req.Method,
		"Query":  req.QueryData,
	}).Debug("Requesting dependant applications list")

	resp, body, errs := req.EndBytes()
	if errs = handleResp(resp, errs); errs != nil {
		return nil, errs
	}

	return body, nil
}

// DependantApplicationUpdate downloads the binary.tar for a specific application and target commit
// Saving it to {ENM_ASSETS_DIRECTORY}/{applicationUUID}/{targetCommit}/binary.tar
func DependantApplicationUpdate(applicationUUID int, targetCommit string) error {
	url, err := buildPath(address, []string{version, "assets", strconv.Itoa(applicationUUID), targetCommit})
	if err != nil {
		return err
	}

	req, err := grab.NewRequest(url)
	if err != nil {
		return err
	}

	q := req.HTTPRequest.URL.Query()
	q.Set("apikey", rawKey)
	req.HTTPRequest.URL.RawQuery = q.Encode()

	filePath := config.GetAssetsDir()
	filePath = path.Join(filePath, strconv.Itoa(applicationUUID))
	filePath = path.Join(filePath, targetCommit)
	if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
		return err
	}
	filePath = path.Join(filePath, "binary.tar")
	req.Filename = filePath

	log.WithFields(log.Fields{
		"URL":         req.HTTPRequest.URL,
		"Method":      req.HTTPRequest.Method,
		"Query":       req.HTTPRequest.URL.RawQuery,
		"Destination": req.Filename,
	}).Debug("Requesting dependant application update")

	client := grab.NewClient()
	_, err = client.Do(req)

	return err
}

// DependantDeviceLog transmits a log message and timestamp for a specific device
func DependantDeviceLog(UUID, message string) []error {
	url, err := buildPath(address, []string{version, "devices", UUID, "logs"})
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

	req := gorequest.New()
	req.Put(url)
	req.Set("Content-Type", "application/json")
	req.Query(key)
	req.Send((string)(bytes))

	log.WithFields(log.Fields{
		"URL":    req.Url,
		"Method": req.Method,
		"Query":  req.QueryData,
		"Body":   (string)(bytes),
	}).Debug("Transmitting dependant device log")

	resp, _, errs := req.End()
	return handleResp(resp, errs)
}

// DependantDeviceInfoUpdate transmits status and is_online for a specific device
func DependantDeviceInfoUpdate(UUID, status string, online bool) []error {
	url, err := buildPath(address, []string{version, "devices", UUID})
	if err != nil {
		return []error{err}
	}

	type dependantDeviceInfo struct {
		Status string `json:"status"`
		Online bool   `json:"is_online"`
	}

	content := &dependantDeviceInfo{
		Status: status,
		Online: online,
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		return []error{err}
	}

	req := gorequest.New()
	req.Put(url)
	req.Set("Content-Type", "application/json")
	req.Query(key)
	req.Send((string)(bytes))

	log.WithFields(log.Fields{
		"URL":    req.Url,
		"Method": req.Method,
		"Query":  req.QueryData,
		"Body":   (string)(bytes),
	}).Debug("Transmitting dependant device info")

	resp, _, errs := req.End()
	return handleResp(resp, errs)
}

// DependantDeviceInfo returns a single dependant device assigned to the edge-node-manager
func DependantDeviceInfo() error {
	return fmt.Errorf("Not implemented")
}

// DependantDeviceProvision provisions a single dependant device to a specific application
func DependantDeviceProvision(applicationUUID int) (string, string, []error) {
	url, err := buildPath(address, []string{version, "devices"})
	if err != nil {
		return "", "", []error{err}
	}

	type dependantDeviceProvision struct {
		ApplicationUUID int `json:"applicationId"` // TEMP: should be appId
	}

	content := &dependantDeviceProvision{
		ApplicationUUID: applicationUUID,
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		return "", "", []error{err}
	}

	req := gorequest.New()
	req.Post(url)
	req.Set("Content-Type", "application/json")
	req.Query(key)
	req.Send((string)(bytes))

	log.WithFields(log.Fields{
		"URL":    req.Url,
		"Method": req.Method,
		"Query":  req.QueryData,
		"Body":   (string)(bytes),
	}).Debug("Requesting dependant device provision")

	resp, body, errs := req.EndBytes()
	if errs = handleResp(resp, errs); errs != nil {
		return "", "", errs
	}

	var buffer map[string]interface{}
	if err := json.Unmarshal(body, &buffer); err != nil {
		return "", "", []error{err}
	}

	return buffer["uuid"].(string), buffer["name"].(string), nil
}

// DependantDevicesList returns all dependant devices assigned to the edge-node-manager
func DependantDevicesList() error {
	return fmt.Errorf("Not implemented")
}

func init() {
	log.SetLevel(config.GetLogLevel())

	address = config.GetSuperAddr()
	version = config.GetSuperAPIVer()
	rawKey = config.GetSuperAPIKey()

	type apiKey struct {
		APIKey string `json:"apikey"`
	}

	content := &apiKey{
		APIKey: rawKey,
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		log.WithFields(log.Fields{
			"Key":   rawKey,
			"Error": err,
		}).Fatal("Unable to marshall API key")
	}
	key = (string)(bytes)

	log.WithFields(log.Fields{
		"Address": address,
		"Version": version,
		"Key":     key,
		"Raw key": rawKey,
	}).Debug("Initialised outgoing supervisor API")
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

	if resp.StatusCode != 200 && resp.StatusCode != 201 { // TODO: pass in accepted response code
		return []error{fmt.Errorf("invalid response received: %s", resp.Status)}
	}

	log.WithFields(log.Fields{
		"Response": resp.Status,
	}).Debug("Valid response received")

	return nil
}
