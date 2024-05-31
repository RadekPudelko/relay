package particle

import (
	"database/sql"
	"fmt"
)

const SomPingError = "|1"
const SomPingOffline = "|2"

const SomCFError = 1
const SomCFBadRV = 2

type MockParticle struct{}

func NewMock() MockParticle {
	return MockParticle{}
}

// Return is decided by the last 2 letters of the som id
func (p MockParticle) Ping(somId string) (bool, error) {
    if len(somId) < 2 {
        return true, nil
    }

	// TODO: add latency
	// TODO: make this random instead
    switch somId[len(somId) - 2:] {
	case SomPingError:
		return false, fmt.Errorf("MockParticle.Ping: error")
	case SomPingOffline:
		return false, nil
	default: // online
		return true, nil
	}
}

// Return is decided by the value returnValue
func (p MockParticle) CloudFunction(somId string, cloudFunction string, argument string, returnValue sql.NullInt64) (bool, error) {
    if !returnValue.Valid {
        return true, nil
    }
	// TODO: add latency
	// TODO: make this random instead
	switch returnValue.Int64 {
	case SomCFError:
		return false, fmt.Errorf("MockParticle.CloudFunction: error")
	case SomCFBadRV:
		return false, nil
	default: // good return value
		return true, nil
	}
}

