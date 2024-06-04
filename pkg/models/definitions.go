package models

import (
    "time"
    "fmt"
)

// TODO: Rename this to something w/o verb
type CreateRelayRequest struct {
	DeviceId          string  `json:"device_id"`
	CloudFunction     string  `json:"cloud_function"`
	Argument          *string `json:"argument,omitempty"`
	DesiredReturnCode *int    `json:"desired_return_code,omitempty"`
	// TODO time comes in a as a string need to parse
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
}

func (p CreateRelayRequest) String() string {
	str := fmt.Sprintf("device: %s, function: %s", p.DeviceId, p.CloudFunction)
	if p.Argument != nil {
		str += fmt.Sprintf(", argument: %s", *p.Argument)
	}
	if p.DesiredReturnCode != nil {
		str += fmt.Sprintf(", desired return code: %d", *p.DesiredReturnCode)
	}
	return str
}
