package mlbstats

import (
	"fmt"
	"time"
)

// Date validates if a Schedule has a given date and returns a DateData object if it exists
func (s Schedule) Date(date time.Time) (DateData, error) {
	if s.Dates == nil {
		return DateData{}, fmt.Errorf("there are no dates in the schedule")
	}

	for _, d := range s.Dates {
		dateString := d.Date
		dateTime, err := time.Parse("2006-01-02", dateString)
		if err != nil {
			continue
		}
		// Check the days are equal
		if dateTime.Year() == date.Year() && dateTime.YearDay() == date.YearDay() {
			return d, nil
		}
	}

	return DateData{}, fmt.Errorf("unable to find a matching date from mlbstats: looking for %v. Received %v", date, s.Dates[0].Date)
}

// InProgress returns a bool given a game's current state if the game is in progress or not
func (s Status) InProgress() bool {
	return (s.DetailedState == "In Progress" || s.DetailedState == "Manager Challenge")
}
