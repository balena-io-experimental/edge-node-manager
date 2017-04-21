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

func removeHotspotConnections(ssid string) error {
	connection, err := dbus.SystemBus()
	if err != nil {
		return err
	}

	var settingsPaths []dbus.ObjectPath
	settingsObject := connection.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager/Settings")
	if err := settingsObject.Call("org.freedesktop.NetworkManager.Settings.ListConnections", 0).Store(&settingsPaths); err != nil {
		return err
	}

	for _, settingsPath := range settingsPaths {
		var settings map[string]map[string]dbus.Variant
		settingsObject := connection.Object("org.freedesktop.NetworkManager", settingsPath)
		if err := settingsObject.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0).Store(&settings); err != nil {
			return err
		}

		if settings["connection"]["id"].Value().(string) == ssid {
			if err := settingsObject.Call("org.freedesktop.NetworkManager.Settings.Connection.Delete", 0).Store(); err != nil {
				return err
			}
		}
	}

	return nil
}

func createHotSpotConnection(ssid, password string) error {
	deviceInterface, devicePath, err := getInterface()
	if err != nil {
		return err
	}

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
	hotspot["connection"]["interface-name"] = deviceInterface
	hotspot["connection"]["type"] = "801-11-wireless"

	hotspot["ipv4"] = make(map[string]interface{})
	hotspot["ipv4"]["method"] = "shared"

	var path, activeConnectionPath dbus.ObjectPath
	rootObject := connection.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	if err := rootObject.Call(
		"org.freedesktop.NetworkManager.AddAndActivateConnection",
		0,
		hotspot,
		devicePath,
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

func getInterface() (string, dbus.ObjectPath, error) {
	connection, err := dbus.SystemBus()
	if err != nil {
		return "", "", err
	}

	var devicesPaths []dbus.ObjectPath
	rootObject := connection.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	if err := rootObject.Call("org.freedesktop.NetworkManager.GetAllDevices", 0).Store(&devicesPaths); err != nil {
		return "", "", err
	}

	for _, devicesPath := range devicesPaths {
		devicesObject := connection.Object("org.freedesktop.NetworkManager", devicesPath)

		value, err := getProperty(devicesObject, "org.freedesktop.NetworkManager.Device.State")
		if err != nil {
			return "", "", err
		}
		deviceState := NmDeviceState(value.(uint32))

		value, err = getProperty(devicesObject, "org.freedesktop.NetworkManager.Device.DeviceType")
		if err != nil {
			return "", "", err
		}
		deviceType := NmDeviceType(value.(uint32))

		if deviceState == NmDeviceStateDisconnected && deviceType == NmDeviceTypeWifi {
			value, err = getProperty(devicesObject, "org.freedesktop.NetworkManager.Device.Interface")
			if err != nil {
				return "", "", err
			}

			return value.(string), devicesPath, nil
		}
	}

	return "", "", fmt.Errorf("No free wifi interface found")
}

func getProperty(object dbus.BusObject, property string) (interface{}, error) {
	value, err := object.GetProperty(property)
	if err != nil {
		return nil, err
	}

	return value.Value(), nil
}
