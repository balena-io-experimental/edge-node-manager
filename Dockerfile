# Debian base-image for the Raspberry Pi 3
# See more about resin base images here: http://docs.resin.io/runtime/resin-base-images/
FROM resin/raspberrypi3-debian:latest

# Disable systemd init system
ENV INITSYSTEM off

# Set our working directory
WORKDIR /usr/src/app

# Add apt source of the foundation repository and install bluez
# We need this source because bluez needs to be patched in order to work with Raspberry Pi 3
RUN apt-get update && apt-get install -yq --no-install-recommends \
    wget && \
    wget http://archive.raspberrypi.org/debian/raspberrypi.gpg.key -O - | apt-key add - && \
    sed -i '1s#^#deb http://archive.raspberrypi.org/debian jessie main\n#' /etc/apt/sources.list && \
    apt-get install -yq --no-install-recommends \
    bluez \
    bluez-firmware && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Use apt-get if you need to install dependencies,
# for instance if you need ALSA sound utils, just uncomment the lines below.
# RUN apt-get update && apt-get install -yq --no-install-recommends \
#    alsa-utils \
#    libasound2-dev && \
#    apt-get clean && rm -rf /var/lib/apt/lists/*

# Copy start script into the working directory
COPY start.sh ./

# Get the edge-node-manager binary, rename and make executable. Ensure you have the correct release
# and architecture by checking https://github.com/resin-io/edge-node-manager/releases/latest
RUN wget https://resin-production-downloads.s3.amazonaws.com/edge-node-manager/v0.1.8/edge-node-manager-v0.1.8-linux-arm && \
    mv edge-node-manager-v0.1.8-linux-arm edge-node-manager && \
    chmod +x edge-node-manager

# Alternatively cross-compile the binary locally (env GOOS=linux GOARCH=arm go build)
# and copy it into the working directory - good for development
# COPY edge-node-manager ./

# start.sh will run when container starts up on the device
CMD ["bash", "start.sh"]
