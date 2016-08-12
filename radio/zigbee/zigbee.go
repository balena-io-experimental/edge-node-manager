package zigbee

import (
	"log"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

/*
Uses the go-serial package
github.com/jacobsa/go-serial/serial
*/

var options serial.OpenOptions

func Configure(port string, baudRate, dataBits, stopBits, minimumReadSize uint) {
	log.Println("Configuring zigbee options")
	options = serial.OpenOptions{
		PortName:        port,
		BaudRate:        baudRate,
		DataBits:        dataBits,
		StopBits:        stopBits,
		MinimumReadSize: minimumReadSize,
	}
}

func Scan(name string, timeout time.Duration) ([]string, error) {
	log.Printf("Scanning for zigbee devices named %s\r\n", name)
	initialise()
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// Make sure to close it later.
	defer port.Close()

	// Write 4 bytes to the port.
	b := []byte{0x00, 0x01, 0x02, 0x03}
	n, err := port.Write(b)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}
	log.Printf("Wrote %d bytes", n)

	return nil, nil
}

func Online(id string, timeout time.Duration) (bool, error) {
	log.Printf("Checking if zigbee device %s is online\r\n", id)
	initialise()
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// Make sure to close it later.
	defer port.Close()

	// Write 4 bytes to the port.
	b := []byte{0x00, 0x01, 0x02, 0x03}
	n, err := port.Write(b)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}
	log.Printf("Wrote %d bytes", n)

	return true, nil
}

func initialise() {
	if options.PortName == "" {
		log.Println("Setting default zigbee options")
		options = serial.OpenOptions{
			PortName:        "/dev/tty.usbserial-A8008HlV",
			BaudRate:        19200,
			DataBits:        8,
			StopBits:        1,
			MinimumReadSize: 4,
		}
	}
}
