package mlbstats

import "time"

// statsAPIScheduleURL returns the URL for all the game schedule data for the given time
func statsAPIScheduleURL(time time.Time) string {
	host := "https://statsapi.mlb.com"
	path := "/api/v1/schedule"
	query := "?language=en&sportId=1&hydrate=game(content(summary,media(epg))),linescore(runners),flags,team,review&date="
	month := time.Format("01")
	day := time.Format("02")
	year := time.Format("2006")
	return host + path + query + month + "/" + day + "/" + year
}
