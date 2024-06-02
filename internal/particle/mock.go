package particle

import (
	"fmt"
)

const DevicePingError = "|1"
const DevicePingOffline = "|2"

const DeviceCFError = 1
const DeviceCFBadRV = 2
const DeviceCFSuccess = 3

type MockParticle struct{}

func NewMock() MockParticle {
	return MockParticle{}
}

// Return is decided by the last 2 letters of the device id
func (p MockParticle) Ping(deviceId string) (bool, error) {
	if len(deviceId) < 2 {
		return true, nil
	}

	// TODO: add latency
	// TODO: make this random instead
	switch deviceId[len(deviceId)-2:] {
	case DevicePingError:
		return false, fmt.Errorf("MockParticle.Ping: error")
	case DevicePingOffline:
		return false, nil
	default: // online
		return true, nil
	}
}

// Return is decided by the value returnValue
func (p MockParticle) CloudFunction(deviceId string, cloudFunction string, argument string, returnValue *int) (bool, error) {
    if returnValue == nil {
		return true, nil
    }
	// TODO: add latency
	// TODO: make this random instead
	switch *returnValue {
	case DeviceCFError:
		return false, fmt.Errorf("MockParticle.CloudFunction: error")
	case DeviceCFBadRV:
		return false, nil
	default: // good return value
		return true, nil
	}
}
