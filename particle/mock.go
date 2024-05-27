package particle

import (
	"database/sql"
	"fmt"
)

type MockParticle struct {}

func New() (*MockParticle, error) {
    return &MockParticle{}, nil
}

func (p *MockParticle) Ping(somId string, productId int) (bool, error) {
    // TODO: add latency
    switch somId {
    case "error":
        return false, fmt.Errorf("MockParticle.Ping: error")
    case "offline":
        return false, nil
    default: // online
        return true, nil
    }
}

func (p *MockParticle) CloudFunction(somId string, productId int, cloudFunction string, argument string, returnValue sql.NullInt64) (bool, error) {
    // TODO: add latency
    switch somId {
    case "error":
        return false, fmt.Errorf("MockParticle.CloudFunction: error")
    case "bad_rv":
        return false, nil
    default: // good return value
        return true, nil
    }
}

