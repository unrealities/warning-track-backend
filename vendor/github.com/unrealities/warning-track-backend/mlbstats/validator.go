package mlbstats

import (
	"fmt"
	"time"
)

// ValidDate validates if a Schedule has a given date
func (s Schedule) ValidDate(date time.Time, dateFmt string) (bool, error) {
	var err error
	games := []Game{}

	for _, d := range s.Dates {
		dateString := d.Date
		dateTime, err := time.Parse(dateFmt, dateString)
		if err != nil {
			break
		}
		if dateTime == date {
			games = d.Games
			break
		}
	}
	if err != nil {
		return false, err
	}
	if len(games) == 0 {
		return false, fmt.Errorf("unable to find a matching date from mlbstats")
	}

	return true, nil
}
