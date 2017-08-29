package status

// Status defines the device statuses
type Status string

const (
	DOWNLOADING Status = "Downloading"
	INSTALLING         = "Installing"
	STARTING           = "Starting"
	STOPPING           = "Stopping"
	IDLE               = "Idle"
)
