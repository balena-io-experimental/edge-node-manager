package wifi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/grandcat/zeroconf"
	"github.com/parnurzeal/gorequest"
	"github.com/resin-io/edge-node-manager/config"
)

var (
	initialised  bool
	avahiTimeout time.Duration
)

type Host struct {
	ip              string
	deviceType      string
	applicationUUID string
	id              string
}

func Initialise() error {
	if initialised {
		return nil
	}

	log.Info("Initialising wifi hotspot")

	os.Setenv("DBUS_SYSTEM_BUS_ADDRESS", "unix:path=/host/run/dbus/system_bus_socket")

	deviceInterface := config.GetHotspotInterface()
	ssid := config.GetHotspotSSID()
	password := config.GetHotspotPassword()

	if err := removeHotspotConnections(ssid); err != nil {
		return err
	}

	var (
		ethernet bool
		device   NmDevice
		err      error
	)

	// If interface environment variable is defined, create the hotspot on that wifi interface
	// If ethernet is connected, create the hotspot on the first wifi interface found
	// If ethernet is not connected, create the hotspot on the first FREE wifi interface found
	if deviceInterface == "" {
		if ethernet, err = isEthernetConnected(); err != nil {
			return err
		} else if ethernet {
			if device, err = getWifiDevice(); err != nil {
				return err
			}
		} else {
			if device, err = getFreeWifiDevice(); err != nil {
				return err
			}
		}
	} else {
		if device, err = getSpecifiedWifiDevice(deviceInterface); err != nil {
			return err
		}
	}

	if err := createHotspotConnection(device, ssid, password); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"SSID":     ssid,
		"Password": password,
		"Device":   device,
	}).Info("Initialised wifi hotspot")

	initialised = true
	return nil
}

func Cleanup() error {
	// Return as we do not want to disable the hotspot
	return nil
}

func Scan(id string) (map[string]struct{}, error) {
	hosts, err := scan()
	if err != nil {
		return nil, err
	}

	online := make(map[string]struct{})
	for _, host := range hosts {
		if host.applicationUUID == id {
			var s struct{}
			online[host.id] = s
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
		if host.id == id {
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
		if host.id == id {
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
	if avahiTimeout, err = config.GetAvahiTimeout(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to load Avahi timeout")
	}

	log.Debug("Initialised wifi")
}

func scan() ([]Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), avahiTimeout)
	defer cancel()

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	var hosts []Host
	go func(entries <-chan *zeroconf.ServiceEntry, hosts *[]Host) {
		for entry := range entries {
			parts := strings.Split(entry.ServiceRecord.Instance, "_")

			if len(entry.AddrIPv4) < 1 || len(parts) < 3 {
				continue
			}

			host := Host{
				ip:              entry.AddrIPv4[0].String(),
				deviceType:      parts[0],
				applicationUUID: parts[1],
				id:              parts[2],
			}
			*hosts = append(*hosts, host)
		}
	}(entries, &hosts)

	err = resolver.Browse(ctx, "_http._tcp", "local", entries)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Unable to scan")
		return nil, err
	}

	<-ctx.Done()

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
