#!/bin/bash

# Enable on-board bluetooth
./enableBle.sh

# Compile and run the edge-node-manager
# The code is recompiled everytime the container restarts which is not ideal but OK for now
echo "Starting edge-node-manager..."
godep go run main.go