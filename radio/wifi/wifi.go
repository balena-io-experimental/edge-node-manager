package wifi

import (
	"log"
	"net/http"
	"time"
)

func Scan(name string, timeout time.Duration) ([]string, error) {
	log.Printf("Scanning for wifi devices named %s\r\n", name)

	http.Get("http://example.com/")

	return nil, nil
}

func Online(id string, timeout time.Duration) (bool, error) {
	log.Printf("Checking if wifi device %s is online\r\n", id)

	http.Get("http://example.com/")

	return true, nil
}

func Post(id, value string) error {
	log.Printf("Checking if wifi device %s is online\r\n", id)

	http.Get("http://example.com/")

	return nil
}
