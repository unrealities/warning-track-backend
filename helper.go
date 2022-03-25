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
	tz, err := time.LoadLocation("PDT")
	if err != nil {
		return time.Time{}, err
	}

	err = json.NewDecoder(reqBody).Decode(&cont)
	// If no body is passed. Return current date in UTC
	switch {
	case err == io.EOF:
		return time.Now().In(tz), nil
	case err != nil:
		return time.Time{}, err
	}

	// If no date key is passed. Return current date in UTC
	if cont.Data.Date == "" {
		return time.Now().In(tz), nil
	}

	return time.Parse(dateFormat, cont.Data.Date)
}
