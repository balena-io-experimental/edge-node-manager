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
	"github.com/parnurzeal/gorequest"
	"github.com/resin-io/edge-node-manager/config"
)

var (
	address string
	version string
	key     string
	rawKey  string
)

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
	if errs = handleResp(resp, errs, 200); errs != nil {
		return nil, errs
	}

	return body, nil
}

// Downloads the binary.tar for a specific application and target commit
// Saving it to {ENM_ASSETS_DIRECTORY}/{applicationUUID}/{targetCommit}/binary.tar
func DependantApplicationUpdate(applicationUUID int, targetCommit string) error {
	url, err := buildPath(address, []string{version, "dependent-apps", strconv.Itoa(applicationUUID), "assets", targetCommit})
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
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp.HTTPResponse.StatusCode != 200 {
		return fmt.Errorf("Dependant application update failed")
	}

	log.Debug("Dependant application update succeeded")

	return nil
}

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
	req.Post(url)
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
	return handleResp(resp, errs, 202)
}

func DependantDeviceInfoUpdateWithOnlineState(UUID, status, commit string, online bool) []error {
	url, err := buildPath(address, []string{version, "devices", UUID})
	if err != nil {
		return []error{err}
	}

	type dependantDeviceInfo struct {
		Status string `json:"status"`
		Online bool   `json:"is_online"`
		Commit string `json:"commit,omitempty"`
	}

	content := &dependantDeviceInfo{
		Status: status,
		Online: online,
		Commit: commit,
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
	return handleResp(resp, errs, 200)
}

func DependantDeviceInfoUpdateWithoutOnlineState(UUID, status, commit string) []error {
	url, err := buildPath(address, []string{version, "devices", UUID})
	if err != nil {
		return []error{err}
	}

	type dependantDeviceInfo struct {
		Status string `json:"status"`
		Commit string `json:"commit,omitempty"`
	}

	content := &dependantDeviceInfo{
		Status: status,
		Commit: commit,
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
	return handleResp(resp, errs, 200)
}

func DependantDeviceInfo(UUID string) ([]byte, []error) {
	url, err := buildPath(address, []string{version, "devices", UUID})
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
	}).Debug("Requesting dependant device info")

	resp, body, errs := req.EndBytes()
	if errs = handleResp(resp, errs, 200); errs != nil {
		return nil, errs
	}

	return body, nil
}

func DependantDeviceProvision(applicationUUID int) (resinUUID, name string, errs []error) {
	url, err := buildPath(address, []string{version, "devices"})
	if err != nil {
		errs = []error{err}
		return
	}

	type dependantDeviceProvision struct {
		ApplicationUUID int `json:"appId"`
	}

	content := &dependantDeviceProvision{
		ApplicationUUID: applicationUUID,
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		errs = []error{err}
		return
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
	if errs = handleResp(resp, errs, 201); errs != nil {
		return
	}

	var buffer map[string]interface{}
	if err := json.Unmarshal(body, &buffer); err != nil {
		errs = []error{err}
		return
	}

	resinUUID = buffer["uuid"].(string)
	name = buffer["name"].(string)

	return
}

func DependantDevicesList() ([]byte, []error) {
	url, err := buildPath(address, []string{version, "devices"})
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
	}).Debug("Requesting dependant devices list")

	resp, body, errs := req.EndBytes()
	if errs = handleResp(resp, errs, 200); errs != nil {
		return nil, errs
	}

	return body, nil
}

func init() {
	log.SetLevel(config.GetLogLevel())

	address = config.GetSuperAddr()
	version = config.GetVersion()
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

func handleResp(resp gorequest.Response, errs []error, statusCode int) []error {
	if errs != nil {
		return errs
	}

	if resp.StatusCode != statusCode {
		return []error{fmt.Errorf("Invalid response received: %s", resp.Status)}
	}

	log.WithFields(log.Fields{
		"Response": resp.Status,
	}).Debug("Valid response received")

	return nil
}
