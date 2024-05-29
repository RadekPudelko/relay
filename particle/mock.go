package particle

import (
	"database/sql"
	"fmt"
)

const SomPingError = "error"
const SomPingOffline = "offline"

const SomCFError = "error"
const SomCFBadRV = "bad_rv"

type MockParticle struct{}

func NewMock() MockParticle {
	return MockParticle{}
}

func (p MockParticle) Ping(somId string) (bool, error) {
	// TODO: add latency
    // TODO: make this random instead
	switch somId {
	case SomPingError:
		return false, fmt.Errorf("MockParticle.Ping: error")
	case SomPingOffline:
		return false, nil
	default: // online
		return true, nil
	}
}

func (p MockParticle) CloudFunction(somId string, cloudFunction string, argument string, returnValue sql.NullInt64) (bool, error) {
	// TODO: add latency
    // TODO: make this random instead
	switch somId {
	case SomCFError:
		return false, fmt.Errorf("MockParticle.CloudFunction: error")
	case SomCFBadRV:
		return false, nil
	default: // good return value
		return true, nil
	}
}

