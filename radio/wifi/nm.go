package wifi

import (
	"fmt"
	"time"

	"github.com/godbus/dbus"
)

type NmDeviceState uint32

const (
	NmDeviceStateUnknown      NmDeviceState = 0
	NmDeviceStateUnmanaged    NmDeviceState = 10
	NmDeviceStateUnavailable  NmDeviceState = 20
	NmDeviceStateDisconnected NmDeviceState = 30
	NmDeviceStatePrepare      NmDeviceState = 40
	NmDeviceStateConfig       NmDeviceState = 50
	NmDeviceStateNeed_auth    NmDeviceState = 60
	NmDeviceStateIp_config    NmDeviceState = 70
	NmDeviceStateIp_check     NmDeviceState = 80
	NmDeviceStateSecondaries  NmDeviceState = 90
	NmDeviceStateActivated    NmDeviceState = 100
	NmDeviceStateDeactivating NmDeviceState = 110
	NmDeviceStateFailed       NmDeviceState = 120
)

type NmDeviceType uint32

const (
	NmDeviceTypeUnknown    NmDeviceType = 0
	NmDeviceTypeEthernet   NmDeviceType = 1
	NmDeviceTypeWifi       NmDeviceType = 2
	NmDeviceTypeUnused1    NmDeviceType = 3
	NmDeviceTypeUnused2    NmDeviceType = 4
	NmDeviceTypeBt         NmDeviceType = 5
	NmDeviceTypeOlpcMesh   NmDeviceType = 6
	NmDeviceTypeWimax      NmDeviceType = 7
	NmDeviceTypeModem      NmDeviceType = 8
	NmDeviceTypeInfiniband NmDeviceType = 9
	NmDeviceTypeBond       NmDeviceType = 10
	NmDeviceTypeVlan       NmDeviceType = 11
	NmDeviceTypeAdsl       NmDeviceType = 12
	NmDeviceTypeBridge     NmDeviceType = 13
	NmDeviceTypeGeneric    NmDeviceType = 14
	NmDeviceTypeTeam       NmDeviceType = 15
)

type NmActiveConnectionState uint32

const (
	NmActiveConnectionStateUnknown      NmActiveConnectionState = 0
	NmActiveConnectionStateActivating   NmActiveConnectionState = 1
	NmActiveConnectionStateActivated    NmActiveConnectionState = 2
	NmActiveConnectionStateDeactivating NmActiveConnectionState = 3
	NmActiveConnectionStateDeactivated  NmActiveConnectionState = 4
)

type NmDevice struct {
	nmPath      dbus.ObjectPath
	nmState     NmDeviceState
	nmType      NmDeviceType
	nmInterface string
}

func removeHotspotConnections(ssid string) error {
	settingsObject, err := getConnection(ssid)
	if err != nil {
		return err
	} else if settingsObject == nil {
		return nil
	}

	if err := settingsObject.Call("org.freedesktop.NetworkManager.Settings.Connection.Delete", 0).Store(); err != nil {
		return err
	}

	for {
		if settingsObject, err := getConnection(ssid); err != nil {
			return err
		} else if settingsObject == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func getConnection(ssid string) (dbus.BusObject, error) {
	connection, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	var settingsPaths []dbus.ObjectPath
	settingsObject := connection.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager/Settings")
	if err := settingsObject.Call("org.freedesktop.NetworkManager.Settings.ListConnections", 0).Store(&settingsPaths); err != nil {
		return nil, err
	}

	for _, settingsPath := range settingsPaths {
		var settings map[string]map[string]dbus.Variant
		settingsObject := connection.Object("org.freedesktop.NetworkManager", settingsPath)
		if err := settingsObject.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0).Store(&settings); err != nil {
			return nil, err
		}

		if settings["connection"]["id"].Value().(string) == ssid {
			return settingsObject, nil
		}
	}

	return nil, nil
}

func isEthernetConnected() (bool, error) {
	devices, err := getDevices()
	if err != nil {
		return false, err
	}

	for _, device := range devices {
		if device.nmType == NmDeviceTypeEthernet && device.nmState == NmDeviceStateActivated {
			return true, nil
		}
	}

	return false, nil
}

func getWifiDevice() (NmDevice, error) {
	devices, err := getDevices()
	if err != nil {
		return NmDevice{}, err
	}

	for _, device := range devices {
		if device.nmType == NmDeviceTypeWifi {
			return device, nil
		}
	}

	return NmDevice{}, fmt.Errorf("No wifi device found")
}

func getFreeWifiDevice() (NmDevice, error) {
	devices, err := getDevices()
	if err != nil {
		return NmDevice{}, err
	}

	for _, device := range devices {
		if device.nmType == NmDeviceTypeWifi && device.nmState == NmDeviceStateDisconnected {
			return device, nil
		}
	}

	return NmDevice{}, fmt.Errorf("No free wifi device found")
}

func createHotspotConnection(device NmDevice, ssid, password string) error {
	connection, err := dbus.SystemBus()
	if err != nil {
		return err
	}

	hotspot := make(map[string]map[string]interface{})

	hotspot["802-11-wireless"] = make(map[string]interface{})
	hotspot["802-11-wireless"]["band"] = "bg"
	hotspot["802-11-wireless"]["hidden"] = false
	hotspot["802-11-wireless"]["mode"] = "ap"
	hotspot["802-11-wireless"]["security"] = "802-11-wireless-security"
	hotspot["802-11-wireless"]["ssid"] = []byte(ssid)

	hotspot["802-11-wireless-security"] = make(map[string]interface{})
	hotspot["802-11-wireless-security"]["key-mgmt"] = "wpa-psk"
	hotspot["802-11-wireless-security"]["psk"] = password

	hotspot["connection"] = make(map[string]interface{})
	hotspot["connection"]["autoconnect"] = false
	hotspot["connection"]["id"] = ssid
	hotspot["connection"]["interface-name"] = device.nmInterface
	hotspot["connection"]["type"] = "801-11-wireless"

	hotspot["ipv4"] = make(map[string]interface{})
	hotspot["ipv4"]["method"] = "shared"

	var path, activeConnectionPath dbus.ObjectPath
	rootObject := connection.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	if err := rootObject.Call(
		"org.freedesktop.NetworkManager.AddAndActivateConnection",
		0,
		hotspot,
		device.nmPath,
		dbus.ObjectPath("/")).
		Store(&path, &activeConnectionPath); err != nil {
		return err
	}

	activeConnectionObject := connection.Object("org.freedesktop.NetworkManager", activeConnectionPath)
	for {
		value, err := getProperty(activeConnectionObject, "org.freedesktop.NetworkManager.Connection.Active.State")
		if err != nil {
			return err
		}

		if NmActiveConnectionState(value.(uint32)) == NmActiveConnectionStateActivated {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func getDevices() ([]NmDevice, error) {
	connection, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	var paths []dbus.ObjectPath
	rootObject := connection.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	if err := rootObject.Call("org.freedesktop.NetworkManager.GetAllDevices", 0).Store(&paths); err != nil {
		return nil, err
	}

	devices := make([]NmDevice, 5)
	for _, path := range paths {
		deviceObject := connection.Object("org.freedesktop.NetworkManager", path)

		device := NmDevice{}
		device.nmPath = path

		value, err := getProperty(deviceObject, "org.freedesktop.NetworkManager.Device.State")
		if err != nil {
			return nil, err
		}
		device.nmState = NmDeviceState(value.(uint32))

		value, err = getProperty(deviceObject, "org.freedesktop.NetworkManager.Device.DeviceType")
		if err != nil {
			return nil, err
		}
		device.nmType = NmDeviceType(value.(uint32))

		value, err = getProperty(deviceObject, "org.freedesktop.NetworkManager.Device.Interface")
		if err != nil {
			return nil, err
		}
		device.nmInterface = value.(string)

		devices = append(devices, device)
	}

	return devices, nil
}

func getProperty(object dbus.BusObject, property string) (interface{}, error) {
	value, err := object.GetProperty(property)
	if err != nil {
		return nil, err
	}

	return value.Value(), nil
}
