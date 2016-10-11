# edge-node-manager
Resin dependent device edge-node-manager written in Go.

## About
The edge-node-manager is an example of a gateway application designed to bridge the gap between Resin OS capable single board 
computers (e.g. the Raspberry Pi) and non Resin OS capable devices (e.g. micro-controllers). It has been designed to make it as
easy as possible to add new supported dependent device types and to run alongside your user application.

The following functionality is implemented:
 - Dependant device detection.
 - Dependant device provisioning.
 - Dependant device restart.
 - Dependent device over-the-air (OTA) updating.
 - Dependent device logging and information updating.

Currently it only supports a single dependent device type, the nRF51822 BLE SoC from Nordic.

## Notes
 - Dependant device support is still in a very early stage of development, there will be bugs!
 - Please file an issue if you encounter any bugs or think of any enhancments you would like to have.
 - Currently only supported on our [staging](https://dashboard.resinstaging.io) site.
 - Only supported in the Resin Supervisor from `v2.5.0` onwards.

## Images
![Alt text](documentation/images/screenshot.png?raw=true "Dashboard screenshot")
![Alt text](documentation/images/setup.jpg?raw=true "Hardware setup")

## Definitions
### Dependant application
A dependent application is a Resin application that targets devices not capable of interacting directly with the Resin API.

The dependent application is scoped under a Resin application, which gets the definition of gateway application.

A dependent application follows the same development cycle as a conventional Resin application:
 - It binds to your git workspace via the `resin remote`.
 - It consists of a Docker application.
 - It offers the same environment and configuration variables management.

There are some key differences:
 - It does not support Dockerfile templating.
 - The Dockerfile must target an x86 base image.
 - The actual firmware must be stored in the `/assets` folder within the built docker image.

### Dependent device
A dependant device is a device not capable of interacting directly with the Resin API - the reasons can be several, the most common are:
 - No direct Internet capabilities.
 - Not able to run the Resin OS (being a microcontroller, for example).

### Gateway application
The gateway application is responsible for detecting, provisioning and managing dependent devices belonging to one of its dependent
applications. This is possible leveraging a new set of endpoints exposed by the [Resin Supervisor](https://github.com/resin-io/resin-supervisor).

The edge-node-manager (this repository) is an example of a gateway application.

### Gateway device
The gateway device runs the gateway application and has the needed on-board radios to communicate with the managed dependant devices, for example:
 - Bluetooth
 - WiFi
 - LoRa
 - ZigBee

Throughout development an RPi3 has been used as the gateway device.

## Guide

### You will need
 - [Raspberry Pi3](https://www.adafruit.com/product/3055)
 - [nRF51822-DK](https://www.nordicsemi.com/eng/Products/nRF51-DK) 

### Getting started - Gateway application
 - Sign up on [resin.io](https://dashboard.resin.io/signup).
 - Work through the [getting started guide](https://docs.resin.io/raspberrypi3/go/getting-started/) and create a new RPi3 gateway 
  application called `micros`.
 - Clone this repository to your local workspace, for example:
```
$ git clone https://github.com/resin-io-playground/edge-node-manager.git
```
 - Add the gateway application `resin remote` to your local workspace using the useful shortcut in the dashboard UI, for example:
```
$ git remote add resin gh_josephroberts@git.resinstaging.io:gh_josephroberts/micros.git
```
 - Push the application to your RPi3, for example:
```
$ git push resin master
```

#### Configuration variables
Configure the application configuration variables using the `Fleet Configuration` tab accessed from the side bar.

Variable Name | Default value | Set value | Description
------------ | ------------- | ------------- | -------------
RESIN_UI_ENABLE_DEPENDENT_APPLICATIONS | `0` | `1` | Enable dependent application support in the UI.
RESIN_SUPERVISOR_DELTA | `0` | `1` | Enable [Delta Updates](https://docs.resin.io/runtime/delta/).
RESIN_DEPENDENT_DEVICES_HOOK_ADDRESS | `http://0.0.0.0:1337/v1/devices/` | `http://127.0.0.1:1337/v1/devices/` | The endpoint used by the Resin Supervisor to communicate with the ENM.

#### Environment variables
Configure the application environment variables using the `Environment Variables` tab accessed from the side bar.

Variable Name | Default value | Set value | Description
------------ | ------------- | ------------- | -------------
ENM_LOG_LEVEL | `Debug` | `Info` | Set the [logging level](https://github.com/Sirupsen/logrus#level-logging).
ENM_CONFIG_LOOP_DELAY | `10` | `10` | Set the time inbetween application processing loops.
ENM_ASSETS_DIRECTORY | `/data/assets` | `/data/assets` | The directory used to store the dependant device firmware.
ENM_DB_DIRECTORY | `/data/database` | `/data/database` | The directory used to store the database.
ENM_DB_NAME| `my.db` | `my.db` | Set the database name.
ENM_API_VERSION | `v1` | `v1` | The supervisor API version.

### Getting started - Dependent application
 - Create a new dependant application called `nRF51822` from the `Dependant Applications` tab accessed from the side bar.
 - Clone the [nRF51822 repository](https://github.com/resin-io-playground/nRF51822) to your local workspace, for example:
```
$ git clone https://github.com/resin-io-playground/nRF51822.git
```
 - Add the dependent application `resin remote` to your local workspace using the useful shortcut in the dashboard UI, for example:
```
$ git remote add resin gh_josephroberts@git.resinstaging.io:gh_josephroberts/nrf51822.git
```
 - Retrive the dependent application `UUID` from the Resin dashboard, for example: If your dependant application URL is
 `https://dashboard.resinstaging.io/apps/13829/devices` the `UUID` is `13829`.
 - Set the dependent application `UUID` on [line 91](https://github.com/resin-io-playground/edge-node-manager/blob/master/application/application.go#L91)
  in the application package to point to your dependant application, for example: `initApplication(13829, micro.NRF51822, radio.BLUETOOTH)`.
  Make sure you add, commit and push the change to the gateway application we set up above.
 - Set the dependent application `UUID` on [line 46](https://github.com/resin-io-playground/nRF51822/blob/master/src/main.c#L46)
 in the main source file to point to your dependant application, for example: `#define DEVICE_NAME "13829"`
 - Add, commit and push the the application to your RPi3, for example:
```
$ git add src/main.c
$ git commit -m "Set the application UUID"
$ git push resin master
```
 - Turn on your nRF51822 dependent device within range of the RPi3 gateway device and watch it appear on the Resin dashboard!

## Dependent devices
### Supported dependant devices
 - nRF51822 - you can find an example dependant application and instructions [here](https://github.com/resin-io-playground/nRF51822).

### Future supported dependant devices
We aim to support the complete [Adafruit Feather](https://www.adafruit.com/categories/835) family.

Other:
 - BBC micro:bit

### Adding support for a dependant device
Coming soon

## Further reading
Coming soon