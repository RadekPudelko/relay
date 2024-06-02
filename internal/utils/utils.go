package utils

import (
	"fmt"
	"time"
)

func ParseDateTime(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05", // full date with time
		"2006-01-02 15:04:05", // full date with time
		"2006-01-02 15:04",    // Date with hours and minutes
		"2006-01-02 15",       // Date with hours
		"2006-01-02",          // Date only
		"2006-01-02T15:04:05", // ISO 8601 with time
		"2006-01-02T15:04",    // ISO 8601 with hours and minutes
		"2006-01-02T15",       // ISO 8601 with hours
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return t, fmt.Errorf("unable to parse date string: %s", dateStr)
}
