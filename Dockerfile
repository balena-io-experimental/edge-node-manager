# Debian base-image for the Raspberry Pi 3
# See more about resin base images here: http://docs.resin.io/runtime/resin-base-images/
FROM resin/raspberrypi3-debian:latest

# Disable systemd init system
ENV INITSYSTEM off

# Set our working directory
WORKDIR /usr/src/app

# Use apt-get if you need to install dependencies,
RUN apt-get update && apt-get install -yq --no-install-recommends \
    bluez \
    bluez-firmware \
    wget && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Copy start script into the working directory
COPY start.sh ./

# Get the edge-node-manager binary, rename and make executable. Ensure you have the correct release
# and architecture by checking https://github.com/resin-io/edge-node-manager/releases/latest
# RUN wget https://resin-production-downloads.s3.amazonaws.com/edge-node-manager/v0.1.9/edge-node-manager-v0.1.9-linux-arm && \
    # mv edge-node-manager-v0.1.9-linux-arm edge-node-manager && \
    # chmod +x edge-node-manager

# Alternatively cross-compile the binary locally (env GOOS=linux GOARCH=arm go build)
# and copy it into the working directory - good for development
COPY edge-node-manager ./

# start.sh will run when container starts up on the device
CMD ["bash", "start.sh"]
