package function

import (
	"encoding/json"
	"io"
	"time"
)

// ParseDate parses the request body and returns a time.Time value of the requested date
func ParseDate(reqBody io.ReadCloser, dateFormat string) (time.Time, error) {
	var d struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(reqBody).Decode(&d); err != nil {
		return time.Time{}, err
	}

	return time.Parse(dateFormat, d.Date)
}
