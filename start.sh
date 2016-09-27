#!/bin/bash

# Enable on-board bluetooth
./enableBle.sh

# Run the edge-node-manager
echo "Starting edge-node-manager..."
./edge-node-manager
