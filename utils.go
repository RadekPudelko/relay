package main

import(
	"database/sql"
    "time"
    "fmt"
)

func parseDateTime(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",  // full date with time
		"2006-01-02 15:04:05",  // full date with time
		"2006-01-02 15:04",     // Date with hours and minutes
		"2006-01-02 15",        // Date with hours
		"2006-01-02",           // Date only
		"2006-01-02T15:04:05",  // ISO 8601 with time
		"2006-01-02T15:04",     // ISO 8601 with hours and minutes
		"2006-01-02T15",        // ISO 8601 with hours
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

func ParseTime(timeStr string) (sql.NullTime, error) {
    var res sql.NullTime
    res.Valid = false
    if timeStr == "" {
        return res, nil
    }
    const layout = "2024-05-17 19:03:38+00:00"
    t, err := time.Parse(layout, timeStr)
    if err != nil {
        const layout = "2024-05-17 19:03:38"
        t, err = time.Parse(layout, timeStr)
        if err != nil {
            return res, fmt.Errorf("Invalid start time format. Use '2024-05-17 19:03:38' or '2024-05-17 19:03:38+00:00'")
        }
    }
    res = sql.NullTime{Valid: true, Time: t}
    return res, nil
}

