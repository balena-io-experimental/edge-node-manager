#!/bin/bash

# Setup DBUS address so it can be used in the container
export DBUS_SYSTEM_BUS_ADDRESS=unix:path=/host/run/dbus/system_bus_socket

echo "Enabling bluetooth..."
if ! /usr/bin/hciattach /dev/ttyAMA0 bcm43xx 921600 noflow -; then
    /usr/bin/hciattach /dev/ttyAMA0 bcm43xx 921600 noflow -
fi

hciconfig hci0 up

if [ `hcitool dev | wc -l` -gt 1 ]; then
    echo "...passed"
    ./edge-node-manager
else
    echo "...failed"
fi
