package push

import "fmt"

var ErrInvalidDevice = fmt.Errorf("Device has been marked invalid by the service provider. It should be removed from the database.")

type DeviceType string

const (
	DeviceTypeIOS     DeviceType = "ios"
	DeviceTypeAndroid DeviceType = "android"
)

type Device struct {
	Token string
	Type  DeviceType
}

type Payload struct {
	Title       string
	Description string
	Data        map[string]interface{}
}

type Feedback struct {
	*Device
	NewToken string
	Error    error
}
