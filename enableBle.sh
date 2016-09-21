#!/bin/bash

FAILED=0

echo "Testing bluetooth on RPI3."

echo "Attaching hci0..."
if ! /usr/bin/hciattach /dev/ttyAMA0 bcm43xx 921600 noflow -; then
    echo "First try failed. Let's try another time."
    /usr/bin/hciattach /dev/ttyAMA0 bcm43xx 921600 noflow -
fi

echo "Bring hci0 up..."
hciconfig hci0 up

echo "Scan for devices..."
if [ `hcitool scan | wc -l` -le 1 ]; then
    FAILED=1
else
    FAILED=0
fi

echo "Test finished."

# Test result
if [ $FAILED -eq 1 ]; then
    echo "TEST FAILED"
else
    echo "TEST PASSED"
fi