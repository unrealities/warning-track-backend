package mlbstats

import (
	"fmt"
	"time"
)

// Date validates if a Schedule has a given date and returns a DateData object if it exists
func (s Schedule) Date(date time.Time) (DateData, error) {
	var parseError error
	for _, d := range s.Dates {
		dateString := d.Date
		dateTime, err := time.Parse("2006-01-02", dateString)
		if err != nil {
			parseError = err
			continue
		}
		// Check the days are equal
		if dateTime.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			return d, nil
		}
	}

	return DateData{}, fmt.Errorf("unable to find a matching date from mlbstats: Looking for %v. Received %v. Error: %e", date, s.Dates[0].Date, parseError)
}

// InProgress returns a bool given a game's current state if the game is in progress or not
func (s Status) InProgress() bool {
	return (s.DetailedState == "In Progress" || s.DetailedState == "Manager Challenge")
}
