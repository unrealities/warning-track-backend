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
			BaseRunnerState: g.Linescore.Offense.BaseRunnerState(),
			Count: Count{
				Balls:   int(g.Linescore.Balls),
				Strikes: int(g.Linescore.Strikes),
			},
			Inning: int(g.Linescore.CurrentInning),
			Li:     0.0,
			Outs:   int(g.Linescore.Outs),
			Score: Score{
				Away: int(g.Linescore.Teams.Away.Runs),
				Home: int(g.Linescore.Teams.Home.Runs),
			},
			State:       0,
			TopOfInning: g.Linescore.IsTopInning,
		}
	}

	return AllSpark{Games}, nil
}
