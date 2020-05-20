package mlbstats

import (
	"fmt"
	"time"
)

// Date validates if a Schedule has a given date and returns a DateData object if it exists
func (s Schedule) Date(date time.Time) (DateData, error) {
	for _, d := range s.Dates {
		dateString := d.Date
		dateTime, err := time.Parse(dateString, "2006-01-02")
		if err != nil {
			continue
		}
		if dateTime == date {
			return d, nil
		}
	}

	return DateData{}, fmt.Errorf("unable to find a matching date from mlbstats")
}

// BaseRunnerState converts individual base statuses to one integer value for easier processing
// 0:empty; 1:1b; 2:2b; 3:3b; 4:1b,2b; 5:1b,3b; 6:2b,3b; 7:1b,2b,3b
func (o Offense) BaseRunnerState() int {
	switch {
	case o.First.ID == 0 && o.Second.ID == 0 && o.Third.ID == 0:
		return 0
	case o.First.ID > 0 && o.Second.ID == 0 && o.Third.ID == 0:
		return 1
	case o.First.ID == 0 && o.Second.ID > 0 && o.Third.ID == 0:
		return 2
	case o.First.ID == 0 && o.Second.ID == 0 && o.Third.ID > 0:
		return 3
	case o.First.ID > 0 && o.Second.ID > 0 && o.Third.ID == 0:
		return 4
	case o.First.ID > 0 && o.Second.ID == 0 && o.Third.ID > 0:
		return 5
	case o.First.ID == 0 && o.Second.ID > 0 && o.Third.ID > 0:
		return 6
	case o.First.ID > 0 && o.Second.ID > 0 && o.Third.ID > 0:
		return 7
	}
	return 0
}

// InProgress returns a bool given a game's current state if the game is in progress or not
func (s Status) InProgress() bool {
	return s.DetailedState == "In Progress"
}
