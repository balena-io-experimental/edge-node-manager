# Define a reusable argument containing the path to the go project
# see: https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact
ARG PROJECT_DIR=/go/src/github.com/resin-io/edge-node-manager

# Build source using an image that has Go and Glide
FROM billyteves/alpine-golang-glide:1.2.0 as GO

ARG PROJECT_DIR

RUN mkdir -p $PROJECT_DIR
COPY . $PROJECT_DIR
WORKDIR $PROJECT_DIR

# Use Glide to install Go dependencies
RUN glide install

# Cross compile the ENM binary
RUN env GOOS=linux GOARCH=arm go build

# Debian base-image
# See more about resin base images here: http://docs.resin.io/runtime/resin-base-images/
FROM resin/%%RESIN_MACHINE_NAME%%-debian

ARG PROJECT_DIR

# Disable systemd init system
ENV INITSYSTEM off

# Set our working directory
WORKDIR /usr/src/app

# The raspberrypi3 requires extra packages - `bluez-firmware`
ENV EXTRA_PACKAGES ""
RUN if [ "%%RESIN_MACHINE_NAME%%" = "raspberrypi3" ]; then export EXTRA_PACKAGES="bluez-firmware"; fi && \
    apt-get update && apt-get install -yq --no-install-recommends \
    $EXTRA_PACKAGES \
    bluez \
    curl \
    jq && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Copy the cross-compiled binary into the working directory
COPY --from=GO $PROJECT_DIR/edge-node-manager ./

# The edge-node-manager will run when container starts up on the device
CMD ["./edge-node-manager"]
