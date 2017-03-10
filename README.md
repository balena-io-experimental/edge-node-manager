# edge-node-manager [![Go Report Card](https://goreportcard.com/badge/github.com/resin-io/edge-node-manager)](https://goreportcard.com/report/github.com/resin-io/edge-node-manager) [![Build Status](https://travis-ci.com/resin-io/edge-node-manager.svg?token=SsmNYChpKvn5yEXMkM2D&branch=master)](https://travis-ci.com/resin-io/edge-node-manager)
resin.io dependent device edge-node-manager written in Go.

## Getting started
 - Sign up on [resin.io](https://dashboard.resin.io/signup)
 - Work through the [getting started guide](https://docs.resin.io/raspberrypi3/nodejs/getting-started/)
 - Create a new Raspberry Pi 3 application called `EdgeNodeManager`
 - Set these variables in the `Fleet Configuration` application side tab
    - `RESIN_SUPERVISOR_DELTA=1`
    - `RESIN_UI_ENABLE_DEPENDENT_APPLICATIONS=1`
 - Clone this repository to your local workspace
 - Add the dependent application `resin remote` to your local workspace
 - Provision a Raspberry Pi 3 gateway device
 - Push code to resin as normal :)
 - Follow the readme of the [supported dependent device](#supported-dependent-devices) you would like to use

## Supported dependent devices
 - [micro:bit](https://github.com/resin-io-projects/micro-bit)
 - [nRF51822-DK](https://github.com/resin-io-projects/nRF51822-DK)

## API
The edge-node-manager provides an API that allows the user to set the target status of the main process and to check whether there are updates pending. This can be useful for a couple of reasons:
 - Ensure the edge-node-manager process is not running during an update of the user container
 - Free up the on-board radios to allow user code to interact directly with the dependent devices e.g. to collect sensor data

**Note** - [Update locks](https://github.com/resin-io/resin-supervisor/blob/master/docs/update-locking.md) *are* enabled whilst the process is running. This means that you will need to pause the edge-node-manager to update the user container.

**Warning** - Do not try and interact with the on-board radios whilst the edge-node-manager is running (this leads to inconsistent, unexpected behaviour).

### SET /v1/enm/status
Set the edge-node-manager process status.

#### Example
```
curl -i -H "Content-Type: application/json" -X PUT --data '{"targetStatus":"Paused"}' localhost:1337/v1/enm/status
curl -i -H "Content-Type: application/json" -X PUT --data '{"targetStatus":"Running"}' localhost:1337/v1/enm/status
```

#### Response
```
HTTP/1.1 200 OK
```

### GET /v1/enm/status
Get the edge-node-manager process status.

#### Example
```
curl -i -X GET localhost:1337/v1/enm/status
```

#### Response
```
HTTP/1.1 200 OK
{
    "currentStatus":"Running",
    "targetStatus":"Paused",
    "updatesPending": true
}
```

## Further reading
### About
The edge-node-manager is an example of a gateway application designed to bridge the gap between Resin OS capable single board
computers (e.g. the Raspberry Pi) and non Resin OS capable devices (e.g. micro-controllers). It has been designed to make it as
easy as possible to add new supported dependent device types and to run alongside your user application.

The following functionality is implemented:
 - Dependent device detection
 - Dependent device provisioning
 - Dependent device restart
 - Dependent device over-the-air (OTA) updating
 - Dependent device logging and information updating
 - API

### Definitions
#### Dependent application
A dependent application is a Resin application that targets devices not capable of interacting directly with the Resin API.

The dependent application is scoped under a Resin application, which gets the definition of gateway application.

A dependent application follows the same development cycle as a conventional Resin application:
 - It binds to your git workspace via the `resin remote`
 - It consists of a Docker application
 - It offers the same environment and configuration variables management

There are some key differences:
 - It does not support Dockerfile templating
 - The Dockerfile must target an x86 base image
 - The actual firmware must be stored in the `/assets` folder within the built docker image

#### Dependent device
A dependent device is a device not capable of interacting directly with the Resin API - the reasons can be several, the most common are:
 - No direct Internet capabilities
 - Not able to run the Resin OS (being a microcontroller, for example)

#### Gateway application
The gateway application is responsible for detecting, provisioning and managing dependent devices belonging to one of its dependent
applications. This is possible leveraging a new set of endpoints exposed by the [Resin Supervisor](https://github.com/resin-io/resin-supervisor).

The edge-node-manager (this repository) is an example of a gateway application.

#### Gateway device
The gateway device runs the gateway application and has the needed on-board radios to communicate with the managed dependent devices, for example:
 - Bluetooth
 - WiFi
 - LoRa
 - ZigBee

Throughout development a Raspberry Pi 3 has been used as the gateway device.
