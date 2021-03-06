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

	if err := json.NewDecoder(reqBody).Decode(&cont); err != nil {
		return time.Time{}, err
	}

	return time.Parse(dateFormat, cont.Data.Date)
}
