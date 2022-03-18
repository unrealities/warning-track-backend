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

	err := json.NewDecoder(reqBody).Decode(&cont)
	// If no date is passed. Return current date in UTC
	switch {
	case err == io.EOF:
		tz, err := time.LoadLocation("PDT")
		if err != nil {
			return time.Time{}, err
		}
		return time.Now().In(tz), nil
	case err != nil:
		return time.Time{}, err
	}

	return time.Parse(dateFormat, cont.Data.Date)
}
