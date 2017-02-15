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
	ble.SetDefaultDevice(device)
	return nil
}

func CloseDevice() error {
	return ble.Stop()
}

func Connect(id string, timeout time.Duration) (ble.Client, error) {
	client, err := ble.Dial(ble.WithSigHandler(context.WithTimeout(context.Background(), timeout*time.Second)), hci.RandomAddress{ble.NewAddr(id)})
	if err != nil {
		return nil, err
	}

	done = make(chan struct{})
	go func() {
		<-client.Disconnected()
		close(done)
	}()

	return client, nil
}

func Disconnect(client ble.Client) error {
	if err := client.CancelConnection(); err != nil {
		return err
	}
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
