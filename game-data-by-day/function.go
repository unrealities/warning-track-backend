package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Date time.Time `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		return
	}

	resp, err := http.Get(statsAPIScheduledURL(d.Date))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	fmt.Println(body)
}

// statsAPIScheduleURL returns the URL for all the game schedule data for the given time
func statsAPIScheduledURL(time time.Time) string {
	host := "http://statsapi.mlb.com"
	path := "/api/v1/schedule"
	query := "?sportId=1&hydrate=game(content(summary,media(epg))),linescore(runners),flags,team&date="
	month := time.Format("01")
	day := time.Format("02")
	year := time.Format("2006")
	return host + path + query + month + "/" + day + "/" + year
}
