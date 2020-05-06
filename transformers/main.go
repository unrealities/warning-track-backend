package transformers

import (
	"time"

	"github.com/unrealities/warning-track-backend/mlbstats"
)

// OptimusPrime takes a mlbstats.Schedule and produces an AllSpark with a day's game data
func OptimusPrime(date time.Time, schedule mlbstats.Schedule) (AllSpark, error) {
	_, err := schedule.Date(date, "2020-02-01")
	if err != nil {
		return AllSpark{}, err
	}

	return AllSpark{}, nil
}
