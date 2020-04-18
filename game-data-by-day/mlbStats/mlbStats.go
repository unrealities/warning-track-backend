package mlbStats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
	"github.com/unrealities/warning-track-backend/game-data-by-day/gCloud"
)

// statsAPIScheduleURL returns the URL for all the game schedule data for the given time
func statsAPIScheduleURL(time time.Time) string {
	host := "http://statsapi.mlb.com"
	path := "/api/v1/schedule"
	query := "?sportId=1&hydrate=game(content(summary,media(epg))),linescore(runners),flags,team&date="
	month := time.Format("01")
	day := time.Format("02")
	year := time.Format("2006")
	return host + path + query + month + "/" + day + "/" + year
}

// GetSchedule returns a Schedule that contains all the requested day's games
func GetSchedule(date time.Time, lg *logging.Logger) (Schedule, error) {
	URL := statsAPIScheduleURL(date)
	lg.Log(logging.Entry{Severity: logging.Debug, Payload: gCloud.LogMessage{Message: fmt.Sprintf("making Get request: %s", URL)}})
	resp, err := http.Get(URL)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gCloud.LogMessage{Message: fmt.Sprintf("error in Get request: %s", err)}})
		return Schedule{}, err
	}
	defer resp.Body.Close()

	lg.Log(logging.Entry{Severity: logging.Debug, Payload: gCloud.LogMessage{Message: "parsing response from Get request"}})
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gCloud.LogMessage{Message: fmt.Sprintf("error reading Get response body: %s", err)}})
		return Schedule{}, err
	}

	lg.Log(logging.Entry{Severity: logging.Debug, Payload: gCloud.LogMessage{Message: "successfully received response from Get"}})

	statsAPIScheduleResp := Schedule{}
	err = json.Unmarshal(body, &statsAPIScheduleResp)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gCloud.LogMessage{Message: fmt.Sprintf("error trying to unmarshal response from statsAPI: %s", err)}})
		return Schedule{}, err
	}

	return statsAPIScheduleResp, nil
}

type Schedule struct {
	Copyright string `json:"copyright"`
	Dates     []struct {
		Date   string        `json:"date"`
		Events []interface{} `json:"events"`
		Games  []struct {
			CalendarEventID string `json:"calendarEventID"`
			Content         struct {
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
			} `json:"content"`
			DayNight     string `json:"dayNight"`
			Description  string `json:"description"`
			DoubleHeader string `json:"doubleHeader"`
			Flags        struct {
				AwayTeamNoHitter    bool `json:"awayTeamNoHitter"`
				AwayTeamPerfectGame bool `json:"awayTeamPerfectGame"`
				HomeTeamNoHitter    bool `json:"homeTeamNoHitter"`
				HomeTeamPerfectGame bool `json:"homeTeamPerfectGame"`
				NoHitter            bool `json:"noHitter"`
				PerfectGame         bool `json:"perfectGame"`
			} `json:"flags"`
			GameDate               string `json:"gameDate"`
			GameNumber             int64  `json:"gameNumber"`
			GamePk                 int64  `json:"gamePk"`
			GameType               string `json:"gameType"`
			GamedayType            string `json:"gamedayType"`
			GamesInSeries          int64  `json:"gamesInSeries"`
			IfNecessary            string `json:"ifNecessary"`
			IfNecessaryDescription string `json:"ifNecessaryDescription"`
			InningBreakLength      int64  `json:"inningBreakLength"`
			Linescore              struct {
				Balls                int64  `json:"balls"`
				CurrentInning        int64  `json:"currentInning"`
				CurrentInningOrdinal string `json:"currentInningOrdinal"`
				Defense              struct {
					Batter struct {
						FullName string `json:"fullName"`
						ID       int64  `json:"id"`
						Link     string `json:"link"`
					} `json:"batter"`
					InHole struct {
						FullName string `json:"fullName"`
						ID       int64  `json:"id"`
						Link     string `json:"link"`
					} `json:"inHole"`
					OnDeck struct {
						FullName string `json:"fullName"`
						ID       int64  `json:"id"`
						Link     string `json:"link"`
					} `json:"onDeck"`
				} `json:"defense"`
				InningHalf  string `json:"inningHalf"`
				InningState string `json:"inningState"`
				Innings     []struct {
					Away struct {
						Errors     int64 `json:"errors"`
						Hits       int64 `json:"hits"`
						LeftOnBase int64 `json:"leftOnBase"`
						Runs       int64 `json:"runs"`
					} `json:"away"`
					Home struct {
						Errors     int64 `json:"errors"`
						Hits       int64 `json:"hits"`
						LeftOnBase int64 `json:"leftOnBase"`
						Runs       int64 `json:"runs"`
					} `json:"home"`
					Num        int64  `json:"num"`
					OrdinalNum string `json:"ordinalNum"`
				} `json:"innings"`
				IsTopInning bool   `json:"isTopInning"`
				Note        string `json:"note"`
				Offense     struct {
					First struct {
						FullName string `json:"fullName"`
						ID       int64  `json:"id"`
						Link     string `json:"link"`
					} `json:"first"`
					Second struct {
						FullName string `json:"fullName"`
						ID       int64  `json:"id"`
						Link     string `json:"link"`
					} `json:"second"`
					Third struct {
						FullName string `json:"fullName"`
						ID       int64  `json:"id"`
						Link     string `json:"link"`
					} `json:"third"`
				} `json:"offense"`
				Outs             int64 `json:"outs"`
				ScheduledInnings int64 `json:"scheduledInnings"`
				Strikes          int64 `json:"strikes"`
				Teams            struct {
					Away struct {
						Errors     int64 `json:"errors"`
						Hits       int64 `json:"hits"`
						LeftOnBase int64 `json:"leftOnBase"`
						Runs       int64 `json:"runs"`
					} `json:"away"`
					Home struct {
						Errors     int64 `json:"errors"`
						Hits       int64 `json:"hits"`
						LeftOnBase int64 `json:"leftOnBase"`
						Runs       int64 `json:"runs"`
					} `json:"home"`
				} `json:"teams"`
			} `json:"linescore"`
			Link              string `json:"link"`
			PublicFacing      bool   `json:"publicFacing"`
			RecordSource      string `json:"recordSource"`
			ScheduledInnings  int64  `json:"scheduledInnings"`
			Season            string `json:"season"`
			SeasonDisplay     string `json:"seasonDisplay"`
			SeriesDescription string `json:"seriesDescription"`
			SeriesGameNumber  int64  `json:"seriesGameNumber"`
			Status            struct {
				AbstractGameCode  string `json:"abstractGameCode"`
				AbstractGameState string `json:"abstractGameState"`
				CodedGameState    string `json:"codedGameState"`
				DetailedState     string `json:"detailedState"`
				StatusCode        string `json:"statusCode"`
			} `json:"status"`
			Teams struct {
				Away struct {
					LeagueRecord struct {
						Losses int64  `json:"losses"`
						Pct    string `json:"pct"`
						Wins   int64  `json:"wins"`
					} `json:"leagueRecord"`
					Score        int64 `json:"score"`
					SeriesNumber int64 `json:"seriesNumber"`
					SplitSquad   bool  `json:"splitSquad"`
					SpringLeague struct {
						Abbreviation string `json:"abbreviation"`
						ID           int64  `json:"id"`
						Link         string `json:"link"`
						Name         string `json:"name"`
					} `json:"springLeague"`
					Team struct {
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
						SpringLeague struct {
							Abbreviation string `json:"abbreviation"`
							ID           int64  `json:"id"`
							Link         string `json:"link"`
							Name         string `json:"name"`
						} `json:"springLeague"`
						TeamCode string `json:"teamCode"`
						TeamName string `json:"teamName"`
						Venue    struct {
							ID   int64  `json:"id"`
							Link string `json:"link"`
							Name string `json:"name"`
						} `json:"venue"`
					} `json:"team"`
				} `json:"away"`
				Home struct {
					LeagueRecord struct {
						Losses int64  `json:"losses"`
						Pct    string `json:"pct"`
						Wins   int64  `json:"wins"`
					} `json:"leagueRecord"`
					Score        int64 `json:"score"`
					SeriesNumber int64 `json:"seriesNumber"`
					SplitSquad   bool  `json:"splitSquad"`
					SpringLeague struct {
						Abbreviation string `json:"abbreviation"`
						ID           int64  `json:"id"`
						Link         string `json:"link"`
						Name         string `json:"name"`
					} `json:"springLeague"`
					Team struct {
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
						SpringLeague struct {
							Abbreviation string `json:"abbreviation"`
							ID           int64  `json:"id"`
							Link         string `json:"link"`
							Name         string `json:"name"`
						} `json:"springLeague"`
						TeamCode string `json:"teamCode"`
						TeamName string `json:"teamName"`
						Venue    struct {
							ID   int64  `json:"id"`
							Link string `json:"link"`
							Name string `json:"name"`
						} `json:"venue"`
					} `json:"team"`
				} `json:"home"`
			} `json:"teams"`
			Tiebreaker string `json:"tiebreaker"`
			Venue      struct {
				ID   int64  `json:"id"`
				Link string `json:"link"`
				Name string `json:"name"`
			} `json:"venue"`
		} `json:"games"`
		TotalEvents          int64 `json:"totalEvents"`
		TotalGames           int64 `json:"totalGames"`
		TotalGamesInProgress int64 `json:"totalGamesInProgress"`
		TotalItems           int64 `json:"totalItems"`
	} `json:"dates"`
	TotalEvents          int64 `json:"totalEvents"`
	TotalGames           int64 `json:"totalGames"`
	TotalGamesInProgress int64 `json:"totalGamesInProgress"`
	TotalItems           int64 `json:"totalItems"`
}
