package mlbstats

// Content is all the mlb game digital content
type Content struct {
	Editorial  struct{} `json:"editorial"`
	GameNotes  struct{} `json:"gameNotes"`
	Highlights struct{} `json:"highlights"`
	Link       string   `json:"link"`
	Media      struct {
		EnhancedGame bool `json:"enhancedGame"`
		Epg          []struct {
			Items []struct {
				CallLetters      string `json:"callLetters"`
				ContentID        string `json:"contentId"`
				Description      string `json:"description"`
				EspnAuthRequired bool   `json:"espnAuthRequired"`
				FoxAuthRequired  bool   `json:"foxAuthRequired"`
				FreeGame         bool   `json:"freeGame"`
				Fs1AuthRequired  bool   `json:"fs1AuthRequired"`
				ID               int64  `json:"id"`
				Language         string `json:"language"`
				MediaFeedSubType string `json:"mediaFeedSubType"`
				MediaFeedType    string `json:"mediaFeedType"`
				MediaID          string `json:"mediaId"`
				MediaState       string `json:"mediaState"`
				MlbnAuthRequired bool   `json:"mlbnAuthRequired"`
				RenditionName    string `json:"renditionName"`
				TbsAuthRequired  bool   `json:"tbsAuthRequired"`
				Type             string `json:"type"`
			} `json:"items"`
			Title string `json:"title"`
		} `json:"epg"`
		EpgAlternate []struct {
			Items []struct {
				Blurb         string `json:"blurb"`
				CclocationVtt string `json:"cclocationVtt"`
				Date          string `json:"date"`
				Description   string `json:"description"`
				Duration      string `json:"duration"`
				Headline      string `json:"headline"`
				ID            string `json:"id"`
				Image         struct {
					AltText interface{} `json:"altText"`
					Cuts    []struct {
						AspectRatio string `json:"aspectRatio"`
						At2x        string `json:"at2x"`
						At3x        string `json:"at3x"`
						Height      int64  `json:"height"`
						Src         string `json:"src"`
						Width       int64  `json:"width"`
					} `json:"cuts"`
					Title string `json:"title"`
				} `json:"image"`
				KeywordsAll []struct {
					DisplayName string `json:"displayName"`
					Type        string `json:"type"`
					Value       string `json:"value"`
				} `json:"keywordsAll"`
				KeywordsDisplay []struct {
					DisplayName string `json:"displayName"`
					Type        string `json:"type"`
					Value       string `json:"value"`
				} `json:"keywordsDisplay"`
				MediaPlaybackID  string `json:"mediaPlaybackId"`
				MediaPlaybackURL string `json:"mediaPlaybackUrl"`
				NoIndex          bool   `json:"noIndex"`
				Playbacks        []struct {
					Height string `json:"height"`
					Name   string `json:"name"`
					URL    string `json:"url"`
					Width  string `json:"width"`
				} `json:"playbacks"`
				SeoTitle string `json:"seoTitle"`
				Slug     string `json:"slug"`
				State    string `json:"state"`
				Title    string `json:"title"`
				Type     string `json:"type"`
			} `json:"items"`
			Title string `json:"title"`
		} `json:"epgAlternate"`
		FreeGame bool `json:"freeGame"`
	} `json:"media"`
	Summary struct {
		HasHighlightsVideo bool `json:"hasHighlightsVideo"`
		HasPreviewArticle  bool `json:"hasPreviewArticle"`
		HasRecapArticle    bool `json:"hasRecapArticle"`
		HasWrapArticle     bool `json:"hasWrapArticle"`
	} `json:"summary"`
}

// DateData is a container for all data for a given day
type DateData struct {
	Date                 string        `json:"date"`
	Events               []interface{} `json:"events"`
	Games                []Game        `json:"games"`
	TotalEvents          int64         `json:"totalEvents"`
	TotalGames           int64         `json:"totalGames"`
	TotalGamesInProgress int64         `json:"totalGamesInProgress"`
	TotalItems           int64         `json:"totalItems"`
}

// Game is all the data for a mlbStats game
type Game struct {
	CalendarEventID string  `json:"calendarEventID"`
	Content         Content `json:"content"`
	DayNight        string  `json:"dayNight"`
	Description     string  `json:"description"`
	DoubleHeader    string  `json:"doubleHeader"`
	Flags           struct {
		AwayTeamNoHitter    bool `json:"awayTeamNoHitter"`
		AwayTeamPerfectGame bool `json:"awayTeamPerfectGame"`
		HomeTeamNoHitter    bool `json:"homeTeamNoHitter"`
		HomeTeamPerfectGame bool `json:"homeTeamPerfectGame"`
		NoHitter            bool `json:"noHitter"`
		PerfectGame         bool `json:"perfectGame"`
	} `json:"flags"`
	GameDate               string       `json:"gameDate"`
	GameNumber             int64        `json:"gameNumber"`
	GamePk                 int64        `json:"gamePk"`
	GameType               string       `json:"gameType"`
	GamedayType            string       `json:"gamedayType"`
	GamesInSeries          int64        `json:"gamesInSeries"`
	IfNecessary            string       `json:"ifNecessary"`
	IfNecessaryDescription string       `json:"ifNecessaryDescription"`
	InningBreakLength      int64        `json:"inningBreakLength"`
	Linescore              Linescore  `json:"linescore"`
	Link                        s   tring    `json:"link"`
	PublicFacing                   bool      `json:"publicFacing"`
	RecordSource                st   ring    `json:"recordSource"`
	ScheduledInnings               int64     `json:"scheduledInnings"`
	Season                      s   tring    `json:"season"`
	SeasonDisplay               st   ring    `json:"seasonDisplay"`
	SeriesDescription           s   tring    `json:"seriesDescription"`
	SeriesGameNumber               int64     `json:"seriesGameNumber"`
	Status                      struct {
		AbstractGameCode  string `json:"abstractGameCode"`
		AbstractGameState string `json:"abstractGameState"`
		CodedGameState    string `json:"codedGameState"`
		DetailedState     string `json:"detailedState"`
		StatusCode        string `json:"statusCode"`
	} `json:"status"`
	Teams struct {
		Away TeamWithRecord `json:"away"`
		Home TeamWithRecord `json:"home"`
	} `json:"teams"`
	Tiebreaker string `json:"tiebreaker"`
	Venue      struct {
		ID   int64  `json:"id"`
		Link string `json:"link"`
		Name string `json:"name"`
	} `json:"venue"`
}

// Linescore is the linescore data for a given game
type Linescore struct {
	Balls                int64  `json:"balls"`
	CurrentInning        int64  `json:"currentInning"`
	CurrentInningOrdinal string `json:"currentInningOrdinal"`
	Defense              struct {
		Batter Player `json:"batter"`
		InHole Player `json:"inHole"`
		OnDeck Player `json:"onDeck"`
	} `json:"defense"`
	InningHalf  string `json:"inningHalf"`
	InningState string `json:"inningState"`
	Innings     []struct {
		Away       TeamLinescore `json:"away"`
		Home       TeamLinescore `json:"home"`
		Num        int64         `json:"num"`
		OrdinalNum string        `json:"ordinalNum"`
	} `json:"innings"`
	IsTopInning bool   `json:"isTopInning"`
	Note        string `json:"note"`
	Offense     struct {
		First  Player `json:"first"`
		Second Player `json:"second"`
		Third  Player `json:"third"`
	} `json:"offense"`
	Outs             int64 `json:"outs"`
	ScheduledInnings int64 `json:"scheduledInnings"`
	Strikes          int64 `json:"strikes"`
	Teams            struct {
		Away TeamLinescore `json:"away"`
		Home TeamLinescore `json:"home"`
	} `json:"teams"`
}

// Player is simple player data
type Player struct {
	FullName string `json:"fullName"`
	ID       int64  `json:"id"`
	Link     string `json:"link"`
}

// Schedule is the format of the json returned from statsAPIScheduleURL
type Schedule struct {
	Copyright            string     `json:"copyright"`
	Dates                []DateData `json:"dates"`
	TotalEvents          int64      `json:"totalEvents"`
	TotalGames           int64      `json:"totalGames"`
	TotalGamesInProgress int64      `json:"totalGamesInProgress"`
	TotalItems           int64      `json:"totalItems"`
}

// SpringLeague is a team's spring league data
type SpringLeague struct {
	Abbreviation string `json:"abbreviation"`
	ID           int64  `json:"id"`
	Link         string `json:"link"`
	Name         string `json:"name"`
}

// Team is data identifying a team of players
type Team struct {
	Abbreviation  string `json:"abbreviation"`
	Active        bool   `json:"active"`
	AllStarStatus string `json:"allStarStatus"`
	Division      struct {
		ID   int64  `json:"id"`
		Link string `json:"link"`
		Name string `json:"name"`
	} `json:"division"`
	FileCode        string `json:"fileCode"`
	FirstYearOfPlay string `json:"firstYearOfPlay"`
	ID              int64  `json:"id"`
	League          struct {
		ID   int64  `json:"id"`
		Link string `json:"link"`
		Name string `json:"name"`
	} `json:"league"`
	Link         string `json:"link"`
	LocationName string `json:"locationName"`
	Name         string `json:"name"`
	Season       int64  `json:"season"`
	ShortName    string `json:"shortName"`
	Sport        struct {
		ID   int64  `json:"id"`
		Link string `json:"link"`
		Name string `json:"name"`
	} `json:"sport"`
	SpringLeague SpringLeague `json:"springLeague"`
	TeamCode     string       `json:"teamCode"`
	TeamName     string       `json:"teamName"`
	Venue        struct {
		ID   int64  `json:"id"`
		Link string `json:"link"`
		Name string `json:"name"`
	} `json:"venue"`
}

// TeamLinescore is team level linescore data
type TeamLinescore struct {
	Errors     int64 `json:"errors"`
	Hits       int64 `json:"hits"`
	LeftOnBase int64 `json:"leftOnBase"`
	Runs       int64 `json:"runs"`
}

// TeamWithRecord is a team with the record, series and spring league information
type TeamWithRecord struct {
	LeagueRecord struct {
		Losses int64  `json:"losses"`
		Pct    string `json:"pct"`
		Wins   int64  `json:"wins"`
	} `json:"leagueRecord"`
	Score        int64        `json:"score"`
	SeriesNumber int64        `json:"seriesNumber"`
	SplitSquad   bool         `json:"splitSquad"`
	SpringLeague SpringLeague `json:"springLeague"`
	Team         Team         `json:"team"`
}
