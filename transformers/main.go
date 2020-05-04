package transformers

import (
	"time"

	"github.com/unrealities/warning-track-backend/mlbstats"
)

// OptimusPrime takes a mlbstats.Schedule and produces an AllSpark with a day's game data
func OptimusPrime(date time.Time, daySchedule mlbstats.Schedule) (AllSpark, error) {
	var err Error
	mlbstatsDateFmt := "2020-02-01"
	mlbstatsGames := []mlbstats.Game{}

	// ensure that we are looping through correct date
	for _, d := range daySchedule.Dates {
		mlbstatsDateString := d.Date
		mlbstatsDate, err := time.Parse(mlbstatsDateFmt, mlbstatsDateString)
		if err != nil {
			break
		}
		if mlbstatsDate == date {
			mlbstatsGames = d.Games
			break
		}
	}
	if err != nil {
		return AllSpark{}, err
	}
	if len(mlbstatsGames) == 0 {
		return AllSpark{}, fmt.Errorf("unable to find a matching date from mlbstats")
	}

	return AllSpark{}, nil
}

// AllSpark contains all the necessary MLB data for Warning-Track to function
// This is a reduced set of data from mlbStats.Schedule
type AllSpark struct {
	Games []Game `json:"games"`
}

// Count holds the game's current at-bat
type Count struct {
	Balls   int `json:"balls"`
	Strikes int `json:"strikes"`
}

// Game holds all the necessary fields of a given game
type Game struct {
	GameTime  time.Time `json:"gameTime"`
	MLBId     int       `json:"mlbID"`
	MLBTVLink string    `json:"mlbTVLink"`
	Status    Status    `json:"status"`
	Teams     Teams     `json:"teams"`
}

// Score holds the game's current score
type Score struct {
	Away int `json:"away"`
	Home int `json:"home"`
}

// Status hold's all the game's current fields. These fields all will change
// during the course of a game
type Status struct {
	BaseRunnerState int     `json:"baseRunnerState"` // 0:none; 1:1b; 2:2b; 3:3b; 4:1b,2b; 5:1b,3b; 6:2b,3b; 7:1b,2b,3b
	Count           Count   `json:"count"`
	Inning          int     `json:"inning"`
	Li              float64 `json:"leverageIndex"`
	Outs            int     `json:"outs"`
	Score           Score   `json:"score"`
	State           int     `json:"state"`
	TopOfInning     bool    `json:"topOfInning"`
}

// Teams holds the teams playing in a given game
type Teams struct {
	AwayID int `json:"away"`
	HomeID int `json:"home"`
}
