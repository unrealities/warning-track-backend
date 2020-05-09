package transformers

import (
	"fmt"
	"time"

	"github.com/unrealities/warning-track-backend/mlbstats"
)

// OptimusPrime takes a mlbstats.Schedule and produces an AllSpark with a day's game data
func OptimusPrime(date time.Time, schedule mlbstats.Schedule) (AllSpark, error) {
	d, err := schedule.Date(date, "2020-02-01")
	if err != nil {
		return AllSpark{}, err
	}

	Games := []Game{}

	for i, g := range d.Games {
		Games[i].MLBId = g.GamePk
		Games[i].MLBTVLink = fmt.Sprintf("https://www.mlb.com/tv/g%v", g.GamePk)

		gameTime, err := time.Parse(g.GameDate, time.RFC3339)
		if err != nil {
			continue
		}
		Games[i].GameTime = gameTime

		Games[i].Teams = Teams{
			AwayID: int(g.Teams.Away.Team.ID),
			HomeID: int(g.Teams.Home.Team.ID),
		}

		// TODO
		Games[i].Status = Status{
			BaseRunnerState: 0,
			Count: Count{
				Balls:   0,
				Strikes: 0,
			},
			Inning: 0,
			Li:     0.0,
			Outs:   0,
			Score: Score{
				Away: 0,
				Home: 0,
			},
			State:       0,
			TopOfInning: true,
		}
	}

	return AllSpark{Games}, nil
}
