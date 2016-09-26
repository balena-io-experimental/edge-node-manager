#!/bin/bash

FAILED=1

echo "Enabling bluetooth."

echo "Attaching hci0..."
if ! /usr/bin/hciattach /dev/ttyAMA0 bcm43xx 921600 noflow -; then
    echo "First try failed. Let's try another time."
    /usr/bin/hciattach /dev/ttyAMA0 bcm43xx 921600 noflow -
fi

echo "Bring hci0 up..."
hciconfig hci0 up

echo "Testing..."
if [ `hcitool dev | wc -l` -gt 1 ]; then
    FAILED=0
fi

if [ $FAILED -eq 1 ]; then
    echo "...failed"
else
    echo "...passed"
fi