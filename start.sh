#!/bin/bash

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
