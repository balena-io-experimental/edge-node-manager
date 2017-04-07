package wifi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lair-framework/go-nmap"
	"github.com/parnurzeal/gorequest"
	"github.com/resin-io/edge-node-manager/config"
)

var wifiDelay time.Duration

type Host struct {
	id  string
	ip  string
	mac string
}

func StartHotspot() error {
	if err := switchConnection("resin-hotspot"); err != nil {
		return err
	}

	// Give the wifi devices a chance to connect
	time.Sleep(wifiDelay)
	return nil
}

func StopHotspot() error {
	return switchConnection("resin-wifi")
}

func Scan(id string) (map[string]struct{}, error) {
	hosts, err := scan()
	if err != nil {
		return nil, err
	}

	online := make(map[string]struct{})
	for _, host := range hosts {
		if host.id == id {
			var s struct{}
			online[host.mac] = s
		}
	}

	return online, nil
}

func Online(id string) (bool, error) {
	hosts, err := scan()
	if err != nil {
		return false, err
	}

	for _, host := range hosts {
		if host.mac == id {
			return true, nil
		}
	}

	return false, nil
}

func GetIP(id string) (string, error) {
	hosts, err := scan()
	if err != nil {
		return "", err
	}

	for _, host := range hosts {
		if host.mac == id {
			return host.ip, nil
		}
	}

	return "", fmt.Errorf("Device offline")
}

func PostForm(url, filePath string) error {
	req := gorequest.New()
	req.Post(url)
	req.Type("multipart")
	req.SendFile(filePath, "firmware.bin", "image")

	log.WithFields(log.Fields{
		"URL":    req.Url,
		"Method": req.Method,
	}).Info("Posting form")

	resp, _, errs := req.End()
	return handleResp(resp, errs, http.StatusOK)
}

func init() {
	log.SetLevel(config.GetLogLevel())

	var err error
	if wifiDelay, err = config.GetWifiDelay(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load wifi delay")
	}

	log.Debug("Initialised wifi")
}

func switchConnection(connection string) error {
	cmd := "python switchConnection.py " + connection
	if err := exec.Command("bash", "-c", cmd).Run(); err != nil {
		log.WithFields(log.Fields{
			"Cmd":   cmd,
			"Error": err,
		}).Error("Unable to switch connection")
		return err
	}

	return nil
}

func scan() ([]Host, error) {
	cmd := exec.Command("bash", "-c", "nmap -sP 10.42.0.* -oX scan.txt")
	if err := cmd.Run(); err != nil {
		log.WithFields(log.Fields{
			"Cmd":   cmd,
			"Error": err,
		}).Error("Unable to scan")
		return nil, err
	}

	file, err := ioutil.ReadFile("scan.txt")
	if err != nil {
		return nil, err
	}

	nmap, err := nmap.Parse(file)
	if err != nil {
		return nil, err
	}

	var hosts []Host
	for _, host := range nmap.Hosts {
		h := Host{}

		for _, address := range host.Addresses {
			if address.AddrType == "mac" {
				h.mac = address.Addr
			} else {
				h.ip = address.Addr
			}
		}

		// Ignore the gateway device
		if h.ip == "10.42.0.1" {
			continue
		}

		url := "http://" + h.ip + "/id"
		resp, body, errs := gorequest.New().Get(url).End()
		if err := handleResp(resp, errs, 200); err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"URL":   url,
				"IP":    h.ip,
				"MAC":   h.mac,
			}).Warn("Unable to get device ID")
			continue
		}
		h.id = body

		hosts = append(hosts, h)
	}

	return hosts, nil
}

func handleResp(resp gorequest.Response, errs []error, statusCode int) error {
	if errs != nil {
		return errs[0]
	}

	if resp.StatusCode != statusCode {
		return fmt.Errorf("Invalid response received: %s", resp.Status)
	}

	log.WithFields(log.Fields{
		"Response": resp.Status,
	}).Debug("Valid response received")

	return nil
}
