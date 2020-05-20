package transformers

import "github.com/unrealities/sabermetrics"

// LeverageIndex uses a game's status and returns a leverage index (float64)
// -1.0 is returned if there is an error
func (s Status) LeverageIndex() float32 {
	baseState := sabermetrics.BaseState{
		First:  s.BaseState.First,
		Second: s.BaseState.Second,
		Third:  s.BaseState.Third,
	}
	score := sabermetrics.Score{
		Away: s.Score.Away,
		Home: s.Score.Home,
	}
	halfInning := sabermetrics.HalfInning{
		Inning:      s.Inning,
		TopOfInning: s.TopOfInning,
	}

	li, err := sabermetrics.LeverageIndex(baseState, score, halfInning, int(s.Outs))
	if err != nil {
		return -1.0
	}
	return li
}
