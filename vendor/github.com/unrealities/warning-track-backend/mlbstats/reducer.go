package mlbstats

import (
	"fmt"
	"time"
)

// Date validates if a Schedule has a given date and returns a DateData object if it exists
func (s Schedule) Date(date time.Time, dateFmt string) (DateData, error) {
	for _, d := range s.Dates {
		dateString := d.Date
		dateTime, err := time.Parse(dateFmt, dateString)
		if err != nil {
			continue
		}
		if dateTime == date {
			return d, nil
		}
	}

	return DateData{}, fmt.Errorf("unable to find a matching date from mlbstats")
}
