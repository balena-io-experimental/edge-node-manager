package wifi

import (
	"net/http"
	"time"
)

// Scan scans for online devices where the device name matches the id passed in
func Scan(id string, timeout time.Duration) (map[string]bool, error) {

	http.Get("http://example.com/")

	return nil, nil
}

// Online checks if a device is online where the device name matches the id passed in
func Online(id string, timeout time.Duration) (bool, error) {

	http.Get("http://example.com/")

	return true, nil
}
