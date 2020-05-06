package transformers

import (
	"time"

	"github.com/unrealities/warning-track-backend/mlbstats"
)

// OptimusPrime takes a mlbstats.Schedule and produces an AllSpark with a day's game data
func OptimusPrime(date time.Time, daySchedule mlbstats.Schedule) (AllSpark, error) {
	dateDate, err := daySchedule.ValidDate(date, "2020-02-01")
	if err != nil {
		return AllSpark{}, err
	}

	return AllSpark{}, nil
}
