package mlbstats

import "time"

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
