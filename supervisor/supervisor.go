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

func WaitUntilReady() {
	log.Info("Waiting until supervisor is ready")

	for {
		resp, _, errs := gorequest.New().Timeout(1 * time.Second).Get(address).End()
		if errs == nil && resp.StatusCode == 401 {
			// The supervisor is up once a 401 status code is returned
			log.Info("Supervisor is ready")
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func DependentApplicationsList() ([]byte, []error) {
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
	}).Debug("Requesting dependent applications list")

	resp, body, errs := req.EndBytes()
	if errs = handleResp(resp, errs, 200); errs != nil {
		return nil, errs
	}

	return body, nil
}

// DependentApplicationUpdate downloads the binary.tar for a specific application and target commit
// Saving it to {ENM_ASSETS_DIRECTORY}/{applicationUUID}/{targetCommit}/binary.tar
func DependentApplicationUpdate(applicationUUID int, targetCommit string) error {
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
	}).Debug("Requesting dependent application update")

	client := grab.NewClient()
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp.HTTPResponse.StatusCode != 200 {
		return fmt.Errorf("Dependent application update failed")
	}

	log.Debug("Dependent application update succeeded")

	return nil
}

func DependentDeviceLog(UUID, message string) []error {
	url, err := buildPath(address, []string{version, "devices", UUID, "logs"})
	if err != nil {
		return []error{err}
	}

	type dependentDeviceLog struct {
		Message   string `json:"message"`
	}

	content := &dependentDeviceLog{
		Message:   message,
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
	}).Debug("Transmitting dependent device log")

	resp, _, errs := req.End()
	return handleResp(resp, errs, 202)
}

func DependentDeviceInfoUpdateWithOnlineState(UUID, status, commit string, online bool) []error {
	url, err := buildPath(address, []string{version, "devices", UUID})
	if err != nil {
		return []error{err}
	}

	type dependentDeviceInfo struct {
		Status string `json:"status"`
		Online bool   `json:"is_online"`
		Commit string `json:"commit,omitempty"`
	}

	content := &dependentDeviceInfo{
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
	}).Warn("Transmitting dependent device info")

	resp, _, errs := req.End()
	return handleResp(resp, errs, 200)
}

func DependentDeviceInfoUpdateWithoutOnlineState(UUID, status, commit string) []error {
	url, err := buildPath(address, []string{version, "devices", UUID})
	if err != nil {
		return []error{err}
	}

	type dependentDeviceInfo struct {
		Status string `json:"status"`
		Commit string `json:"commit,omitempty"`
	}

	content := &dependentDeviceInfo{
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
	}).Warn("Transmitting dependent device info")

	resp, _, errs := req.End()
	return handleResp(resp, errs, 200)
}

func DependentDeviceInfo(UUID string) ([]byte, []error) {
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
	}).Debug("Requesting dependent device info")

	resp, body, errs := req.EndBytes()
	if errs = handleResp(resp, errs, 200); errs != nil {
		return nil, errs
	}

	return body, nil
}

func DependentDeviceProvision(applicationUUID int) (resinUUID, name string, errs []error) {
	url, err := buildPath(address, []string{version, "devices"})
	if err != nil {
		errs = []error{err}
		return
	}

	type dependentDeviceProvision struct {
		ApplicationUUID int `json:"appId"`
	}

	content := &dependentDeviceProvision{
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
	}).Debug("Requesting dependent device provision")

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

func DependentDevicesList() ([]byte, []error) {
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
	}).Debug("Requesting dependent devices list")

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

	// Allow 404 and 410 here as it means the dep. app or dep. device has just been deleted
	if resp.StatusCode != statusCode && resp.StatusCode != 404 && resp.StatusCode != 410 {
		return []error{fmt.Errorf("Invalid response received: %s", resp.Status)}
	}

	log.WithFields(log.Fields{
		"Response": resp.Status,
	}).Debug("Valid response received")

	return nil
}
