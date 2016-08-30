package wifi

import (
	"net/http"
	"time"
)

func Scan(name string, timeout time.Duration) (map[string]bool, error) {

	http.Get("http://example.com/")

	return nil, nil
}

func Online(id string, timeout time.Duration) (bool, error) {

	http.Get("http://example.com/")

	return true, nil
}

func Post(id, value string) error {

	http.Get("http://example.com/")

	return nil
}
