package board

type Type string

const (
	MICROBIT   Type = "microbit"
	NRF51822DK      = "nrf51822dk"
	ESP8266         = "esp8266"
)

type Interface interface {
	InitialiseRadio() error
	CleanupRadio() error
	Update(filePath string) error
	Scan(applicationUUID int) (map[string]struct{}, error)
	Online() (bool, error)
	Restart() error
	Identify() error
	UpdateConfig(interface{}) error
	UpdateEnvironment(interface{}) error
}
