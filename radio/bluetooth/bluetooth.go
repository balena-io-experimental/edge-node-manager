package bluetooth

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
	"github.com/currantlabs/ble/linux/hci"
	"github.com/currantlabs/ble/linux/hci/cmd"
	"github.com/resin-io/edge-node-manager/config"
)

var (
	done chan struct{}
	name *ble.Characteristic
)

func OpenDevice() error {
	device, err := linux.NewDevice()
	if err != nil {
		return err
	}
	updateLinuxParam(device)
	ble.SetDefaultDevice(device)
	return nil
}

func CloseDevice() error {
	return ble.Stop()
}

func ResetDevice() error {
	if err := ble.Stop(); err != nil {
		return err
	}
	return OpenDevice()
}

func Connect(id string, timeout time.Duration) (ble.Client, error) {
	client, err := ble.Dial(ble.WithSigHandler(context.WithTimeout(context.Background(), timeout*time.Second)), hci.RandomAddress{ble.NewAddr(id)})
	if err != nil {
		return nil, err
	}

	// Try 23 here if 500 does not work
	if _, err := client.ExchangeMTU(500); err != nil {
		return nil, err
	}

	done = make(chan struct{})
	go func() {
		<-client.Disconnected()
		close(done)
		log.Info("Closed connection")
	}()

	log.Info("Opened connection")
	return client, nil
}

func Disconnect(client ble.Client) error {
	if err := client.ClearSubscriptions(); err != nil {
		return err
	}

	if err := client.CancelConnection(); err != nil {
		return err
	}
	log.Info("Closing connection")
	<-done
	return nil
}

func Scan(id string, timeout time.Duration) (map[string]bool, error) {
	devices := make(map[string]bool)
	advChannel := make(chan ble.Advertisement)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), timeout*time.Second))

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case adv := <-advChannel:
				if strings.EqualFold(adv.LocalName(), id) {
					devices[adv.Address().String()] = true
				}
			}
		}
	}()

	err := ble.Scan(ctx, false, func(adv ble.Advertisement) { advChannel <- adv }, nil)
	if errors.Cause(err) != context.DeadlineExceeded && errors.Cause(err) != context.Canceled {
		return devices, err
	}

	return devices, nil
}

func Online(id string, timeout time.Duration) (bool, error) {
	online := false
	advChannel := make(chan ble.Advertisement)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = ble.WithSigHandler(context.WithTimeout(ctx, timeout*time.Second))

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case adv := <-advChannel:
				if strings.EqualFold(adv.Address().String(), id) {
					online = true
					cancel()
				}
			}
		}
	}()

	err := ble.Scan(ctx, false, func(adv ble.Advertisement) { advChannel <- adv }, nil)
	if errors.Cause(err) != context.DeadlineExceeded && errors.Cause(err) != context.Canceled {
		return online, err
	}

	return online, nil
}

func GetName(id string, timeout time.Duration) (string, error) {
	client, err := Connect(id, timeout)
	if err != nil {
		return "", err
	}

	resp, err := client.ReadCharacteristic(name)
	if err != nil {
		return "", err
	}

	if err := Disconnect(client); err != nil {
		return "", err
	}

	return string(resp), nil
}

func GetCharacteristic(uuid string, property ble.Property, handle, vhandle uint16) (*ble.Characteristic, error) {
	parsedUUID, err := ble.Parse(uuid)
	if err != nil {
		return nil, err
	}

	characteristic := ble.NewCharacteristic(parsedUUID)
	characteristic.Property = property
	characteristic.Handle = handle
	characteristic.ValueHandle = vhandle

	return characteristic, nil
}

func GetDescriptor(uuid string, handle uint16) (*ble.Descriptor, error) {
	parsedUUID, err := ble.Parse(uuid)
	if err != nil {
		return nil, err
	}

	descriptor := ble.NewDescriptor(parsedUUID)
	descriptor.Handle = handle

	return descriptor, nil
}

func init() {
	log.SetLevel(config.GetLogLevel())

	if err := OpenDevice(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Unable to create a new ble device")
	}

	var err error
	name, err = GetCharacteristic("2a00", ble.CharRead+ble.CharWrite, 0x02, 0x03)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Created a new ble device")
}

func updateLinuxParam(device *linux.Device) error {
	// if err := device.HCI.Send(&cmd.LESetScanParameters{
	//     LEScanType:           0x00,   // 0x00: passive, 0x01: active
	//     LEScanInterval:       0x0062, // 0x0004 - 0x4000; N * 0.625msec
	//     LEScanWindow:         0x0030, // 0x0004 - 0x4000; N * 0.625msec
	//     OwnAddressType:       0x00,   // 0x00: public, 0x01: random
	//     ScanningFilterPolicy: 0x01,   // 0x00: accept all, 0x01: ignore non-white-listed.
	// }, nil); err != nil {
	//     return errors.Wrap(err, "can't set scan param")
	// }

	if err := device.HCI.Option(hci.OptConnParams(
		cmd.LECreateConnection{
			LEScanInterval:        0x0060,    // 0x0004 - 0x4000; N * 0.625 msec
			LEScanWindow:          0x0060,    // 0x0004 - 0x4000; N * 0.625 msec
			InitiatorFilterPolicy: 0x00,      // White list is not used
			PeerAddressType:       0x00,      // Public Device Address
			PeerAddress:           [6]byte{}, //
			OwnAddressType:        0x00,      // Public Device Address
			ConnIntervalMin:       0x0028,    // 0x0006 - 0x0C80; N * 1.25 msec
			ConnIntervalMax:       0x0038,    // 0x0006 - 0x0C80; N * 1.25 msec
			ConnLatency:           0x0000,    // 0x0000 - 0x01F3; N * 1.25 msec
			SupervisionTimeout:    0x002A,    // 0x000A - 0x0C80; N * 10 msec
			MinimumCELength:       0x0000,    // 0x0000 - 0xFFFF; N * 0.625 msec
			MaximumCELength:       0x0000,    // 0x0000 - 0xFFFF; N * 0.625 msec
		})); err != nil {
		return errors.Wrap(err, "can't set connection param")
	}
	return nil
}
