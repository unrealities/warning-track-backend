package sabermetrics

// Int converts a BaseState to one of the following ints
//   0: baseState{first: false, second: false, third: false}
//   1: baseState{first: true, second: false, third: false}
//   2: baseState{first: false, second: true, third: false}
//   3: baseState{first: false, second: false, third: true}
//   4: baseState{first: true, second: true, third: false}
//   5: baseState{first: true, second: false, third: true}
//   6: baseState{first: false, second: true, third: true}
//   7: baseState{first: true, second: true, third: true}
//
// This can be useful for referencing matrices of data like in LeverageIndex
func (bs BaseState) Int() int {
	switch {
	case !bs.First && !bs.Second && !bs.Third:
		return 0
	case bs.First && !bs.Second && !bs.Third:
		return 1
	case !bs.First && bs.Second && !bs.Third:
		return 2
	case !bs.First && !bs.Second && bs.Third:
		return 3
	case bs.First && bs.Second && !bs.Third:
		return 4
	case bs.First && !bs.Second && bs.Third:
		return 5
	case !bs.First && bs.Second && bs.Third:
		return 6
	case bs.First && bs.Second && bs.Third:
		return 7
	}
	return 0
}

// Int converts a HalfInning to an integer value
//   0 : Top of the 1st
//   1 : Bottom of the 1st
//   2 : Top of the 2nd
//   ...
//   17 : Bottom of the 9th
//   ...
func (h HalfInning) Int() int {
	if h.TopOfInning {
		return 2*h.Inning - 2
	}
	return 2*h.Inning - 1
}
