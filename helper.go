package function

import (
	"encoding/json"
	"io"
	"time"
)

// ParseDate parses the request body and returns a time.Time value of the requested date
func ParseDate(reqBody io.ReadCloser, dateFormat string) (time.Time, error) {
	type d struct {
		Date string `json:"date"`
	}
	type data struct {
		Data d `json:"data"`
	}
	var cont data

	// Default if date cannot be determined
	tz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.Time{}, err
	}
	defaultDate := time.Now().In(tz)

	err = json.NewDecoder(reqBody).Decode(&cont)
	if err != nil {
		return defaultDate, nil
	}
	if cont.Data.Date == "" {
		return defaultDate, nil
	}

	return time.Parse(dateFormat, cont.Data.Date)
}
