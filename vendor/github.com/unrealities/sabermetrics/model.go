package sabermetrics

// BaseState represents the current offensive base state
// Each base is given a `bool` to tell if it is occupied or not
type BaseState struct {
	First  bool
	Second bool
	Third  bool
}

// Score represents the current score of a game
type Score struct {
	Away int
	Home int
}

// HalfInning represents the current half inning of a game
type HalfInning struct {
	Inning      int
	TopOfInning bool
}
