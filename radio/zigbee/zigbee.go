package zigbee

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/jacobsa/go-serial/serial"
)

/*
 * Uses the go-serial package
 * github.com/jacobsa/go-serial/serial
 */

var options serial.OpenOptions

func Configure(port string, baudRate, dataBits, stopBits, minimumReadSize uint) {
	options = serial.OpenOptions{
		PortName:        port,
		BaudRate:        baudRate,
		DataBits:        dataBits,
		StopBits:        stopBits,
		MinimumReadSize: minimumReadSize,
	}

	log.WithFields(log.Fields{
		"Options": options,
	}).Debug("Configured ZigBee options")
}

func Scan(name string, timeout time.Duration) (map[string]bool, error) {
	return nil, nil
}

func Online(id string, timeout time.Duration) (bool, error) {
	return true, nil
}

func init() {
	options = serial.OpenOptions{
		PortName:        "/dev/tty.usbserial-A8008HlV",
		BaudRate:        19200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	log.WithFields(log.Fields{
		"Options": options,
	}).Debug("Set default ZigBee options")
}
