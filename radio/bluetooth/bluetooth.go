package bluetooth

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

/*
 * Uses the gatt package
 * https://github.com/paypal/gatt
 */
var (
	Radio         gatt.Device
	periphChannel = make(chan gatt.Peripheral)
)

func Scan(id string, timeout time.Duration) (map[string]bool, error) {
	Radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	if err := Radio.Init(onStateChanged); err != nil {
		return nil, err
	}

	devices := make(map[string]bool)

	for {
		select {
		case <-time.After(timeout * time.Second):
			Radio.StopScanning()
			return devices, nil
		case onlineDevice := <-periphChannel:
			if onlineDevice.Name() == id {
				devices[onlineDevice.ID()] = true
			}
		}
	}
}

func Online(id string, timeout time.Duration) (bool, error) {
	Radio.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	if err := Radio.Init(onStateChanged); err != nil {
		return false, err
	}

	for {
		select {
		case <-time.After(timeout * time.Second):
			Radio.StopScanning()
			return false, nil
		case onlineDevice := <-periphChannel:
			if onlineDevice.ID() == id {
				Radio.StopScanning()
				return true, nil
			}
		}
	}
}

func init() {
	var err error
	if Radio, err = gatt.NewDevice(option.DefaultClientOptions...); err != nil {
		log.WithFields(log.Fields{
			"Options": option.DefaultClientOptions,
			"Error":   err,
		}).Fatal("Unable to create a new gatt device")
	}

	log.WithFields(log.Fields{
		"Options": option.DefaultClientOptions,
	}).Debug("Created new gatt device")
}

func onStateChanged(device gatt.Device, state gatt.State) {
	switch state {
	case gatt.StatePoweredOn:
		device.Scan([]gatt.UUID{}, false)
	default:
		device.StopScanning()
	}
}

func onPeriphDiscovered(peripheral gatt.Peripheral, advertisment *gatt.Advertisement, rssi int) {
	periphChannel <- peripheral
}

// [EE:50:F0:F8:3C:FF][LE]> help
// help                                           Show this help
// exit                                           Exit interactive mode
// quit                                           Exit interactive mode
// connect         [address [address type]]       Connect to a remote device
// disconnect                                     Disconnect from a remote device
// primary         [UUID]                         Primary Service Discovery
// included        [start hnd [end hnd]]          Find Included Services
// characteristics [start hnd [end hnd [UUID]]]   Characteristics Discovery
// char-desc       [start hnd] [end hnd]          Characteristics Descriptor Discovery
// char-read-hnd   <handle>                       Characteristics Value/Descriptor Read by handle
// char-read-uuid  <UUID> [start hnd] [end hnd]   Characteristics Value/Descriptor Read by UUID
// char-write-req  <handle> <new value>           Characteristic Value Write (Write Request)
// char-write-cmd  <handle> <new value>           Characteristic Value Write (No response)
// sec-level       [low | medium | high]          Set security level. Default: low
// mtu             <value>                        Exchange MTU for GATT/ATT
// [EE:50:F0:F8:3C:FF][LE]> primary
// Command Failed: Disconnected
// [EE:50:F0:F8:3C:FF][LE]> connect
// Attempting to connect to EE:50:F0:F8:3C:FF
// Connection successful
// [EE:50:F0:F8:3C:FF][LE]> primary
// attr handle: 0x0001, end grp handle: 0x0007 uuid: 00001800-0000-1000-8000-00805f9b34fb
// attr handle: 0x0008, end grp handle: 0x000b uuid: 00001801-0000-1000-8000-00805f9b34fb
// attr handle: 0x000c, end grp handle: 0x0013 uuid: 00001530-1212-efde-1523-785feabcd123
// attr handle: 0x0014, end grp handle: 0xffff uuid: 0000f00d-1212-efde-1523-785fef13d123
// [EE:50:F0:F8:3C:FF][LE]> characteristics
// handle: 0x0002, char properties: 0x0a, char value handle: 0x0003, uuid: 00002a00-0000-1000-8000-00805f9b34fb
// handle: 0x0004, char properties: 0x02, char value handle: 0x0005, uuid: 00002a01-0000-1000-8000-00805f9b34fb
// handle: 0x0006, char properties: 0x02, char value handle: 0x0007, uuid: 00002a04-0000-1000-8000-00805f9b34fb
// handle: 0x0009, char properties: 0x20, char value handle: 0x000a, uuid: 00002a05-0000-1000-8000-00805f9b34fb
// handle: 0x000d, char properties: 0x04, char value handle: 0x000e, uuid: 00001532-1212-efde-1523-785feabcd123
// handle: 0x000f, char properties: 0x18, char value handle: 0x0010, uuid: 00001531-1212-efde-1523-785feabcd123
// handle: 0x0012, char properties: 0x02, char value handle: 0x0013, uuid: 00001534-1212-efde-1523-785feabcd123
// handle: 0x0015, char properties: 0x08, char value handle: 0x0016, uuid: 0000beef-1212-efde-1523-785fef13d123
// handle: 0x0017, char properties: 0x08, char value handle: 0x0018, uuid: 0000feed-1212-efde-1523-785fef13d123
// [EE:50:F0:F8:3C:FF][LE]> char-desc
// handle: 0x0001, uuid: 00002800-0000-1000-8000-00805f9b34fb
// handle: 0x0002, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x0003, uuid: 00002a00-0000-1000-8000-00805f9b34fb
// handle: 0x0004, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x0005, uuid: 00002a01-0000-1000-8000-00805f9b34fb
// handle: 0x0006, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x0007, uuid: 00002a04-0000-1000-8000-00805f9b34fb
// handle: 0x0008, uuid: 00002800-0000-1000-8000-00805f9b34fb
// handle: 0x0009, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x000a, uuid: 00002a05-0000-1000-8000-00805f9b34fb
// handle: 0x000b, uuid: 00002902-0000-1000-8000-00805f9b34fb
// handle: 0x000c, uuid: 00002800-0000-1000-8000-00805f9b34fb
// handle: 0x000d, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x000e, uuid: 00001532-1212-efde-1523-785feabcd123
// handle: 0x000f, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x0010, uuid: 00001531-1212-efde-1523-785feabcd123
// handle: 0x0011, uuid: 00002902-0000-1000-8000-00805f9b34fb
// handle: 0x0012, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x0013, uuid: 00001534-1212-efde-1523-785feabcd123
// handle: 0x0014, uuid: 00002800-0000-1000-8000-00805f9b34fb
// handle: 0x0015, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x0016, uuid: 0000beef-1212-efde-1523-785fef13d123
// handle: 0x0017, uuid: 00002803-0000-1000-8000-00805f9b34fb
// handle: 0x0018, uuid: 0000feed-1212-efde-1523-785fef13d123

// ➜  examples git:(master) ✗ ./explorer EE:50:F0:F8:3C:FF
// 2016/08/31 17:24:34 dev: hci0 up
// 2016/08/31 17:24:34 dev: hci0 down
// 2016/08/31 17:24:34 dev: hci0 opened
// State: PoweredOn
// Scanning...

// Peripheral ID:EE:50:F0:F8:3C:FF, NAME:(resin)
//   Local Name        = resin
//   TX Power Level    = 0
//   Manufacturer Data = []
//   Service Data      = []

// Connected
// Service: 1800 (Generic Access)
//   Characteristic  2a00 (Device Name)
//     properties    read write
//     value         726573696e | "resin"
//   Characteristic  2a01 (Appearance)
//     properties    read
//     value         0000 | "\x00\x00"
//   Characteristic  2a04 (Peripheral Preferred Connection Parameters)
//     properties    read
//     value         0600200000009001 | "\x06\x00 \x00\x00\x00\x90\x01"

// Service: 1801 (Generic Attribute)
//   Characteristic  2a05 (Service Changed)
//     properties    indicate
//   Descriptor      2902 (Client Characteristic Configuration)
//     value         0000 | "\x00\x00"

// Service: 000015301212efde1523785feabcd123
//   Characteristic  000015321212efde1523785feabcd123
//     properties    writeWithoutResponse
//   Characteristic  000015311212efde1523785feabcd123
//     properties    write notify
//   Descriptor      2902 (Client Characteristic Configuration)
//     value         0000 | "\x00\x00"
//   Characteristic  000015341212efde1523785feabcd123
//     properties    read
//     value         0100 | "\x01\x00"

// Service: 0000f00d1212efde1523785fef13d123
//   Characteristic  0000beef1212efde1523785fef13d123
//     properties    write
//   Characteristic  0000feed1212efde1523785fef13d123
//     properties    write

// Waiting for 5 seconds to get some notifiations, if any.
// Disconnected
// Done
