package transformers

import "time"

// AllSpark contains all the necessary MLB data for Warning-Track to function
// This is a reduced set of data from mlbStats.Schedule
type AllSpark struct {
	Games []Game `json:"games"`
}

// BaseState is a simple vision of the base runner status
type BaseState struct {
	First  bool
	Second bool
	Third  bool
}

// Count holds the game's current at-bat
type Count struct {
	Balls   int `json:"balls"`
	Strikes int `json:"strikes"`
}

// Game holds all the necessary fields of a given game
type Game struct {
	GameTime      time.Time `json:"gameTime"`
	LeverageIndex float32   `json:"leverageIndex"`
	MLBId         int64     `json:"mlbID"`
	MLBTVLink     string    `json:"mlbTVLink"`
	Status        Status    `json:"status"`
	Teams         Teams     `json:"teams"`
}

// Score holds the game's current score
type Score struct {
	Away int `json:"away"`
	Home int `json:"home"`
}

// Status hold's all the game's current fields. These fields all will change
// during the course of a game
type Status struct {
	BaseState   BaseState `json:"baseState"`
	Count       Count     `json:"count"`
	Inning      int       `json:"inning"`
	InProgress  bool      `json:"inProgress"`
	Outs        int       `json:"outs"`
	Score       Score     `json:"score"`
	TopOfInning bool      `json:"topOfInning"`
}

// Teams holds the teams playing in a given game
type Teams struct {
	AwayID int `json:"away"`
	HomeID int `json:"home"`
}
