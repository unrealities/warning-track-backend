package transformers

import (
	"fmt"
	"time"

	"github.com/unrealities/warning-track-backend/mlbstats"
)

// OptimusPrime takes a mlbstats.Schedule and produces an AllSpark with a day's game data
func OptimusPrime(date time.Time, schedule mlbstats.Schedule) (AllSpark, error) {
	d, err := schedule.Date(date)
	if err != nil {
		return AllSpark{}, err
	}

	Games := make([]Game, len(d.Games))

	for i, g := range d.Games {
		Games[i].MLBId = g.GamePk
		Games[i].MLBTVLink = fmt.Sprintf("https://www.mlb.com/tv/g%v", g.GamePk)

		gameTime, err := time.Parse(time.RFC3339, g.GameDate)
		if err != nil {
			continue
		}
		Games[i].GameTime = gameTime

		Games[i].Teams = Teams{
			AwayID: int(g.Teams.Away.Team.ID),
			HomeID: int(g.Teams.Home.Team.ID),
		}

		Games[i].Status = Status{
			BaseState: BaseState{
				First:  g.Linescore.Offense.First.ID > 0,
				Second: g.Linescore.Offense.Second.ID > 0,
				Third:  g.Linescore.Offense.Third.ID > 0,
			},
			Count: Count{
				Balls:   int(g.Linescore.Balls),
				Strikes: int(g.Linescore.Strikes),
			},
			Inning:     int(g.Linescore.CurrentInning),
			InProgress: g.Status.InProgress(),
			Outs:       int(g.Linescore.Outs),
			Score: Score{
				Away: int(g.Linescore.Teams.Away.Runs),
				Home: int(g.Linescore.Teams.Home.Runs),
			},
			TopOfInning: g.Linescore.IsTopInning,
		}

		Games[i].LeverageIndex = Games[i].Status.LeverageIndex()
	}

	return AllSpark{Games}, nil
}
