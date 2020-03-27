package function

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Date time.Time `json:"date"`
	}
	log.Printf("Received request: %+v", r.Body)

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		log.Printf("Error attempting to decode json body: %s", err)
		return
	}
	log.Printf("Date requested: %+v", d.Date)

	URL := statsAPIScheduledURL(d.Date)
	log.Printf("Making Get request: %s", URL)
	resp, err := http.Get(URL)
	if err != nil {
		log.Printf("Error in Get request: %s", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Parsing response from Get request: %+v", resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Get response body: %s", err)
		return
	}

	log.Printf("successfully received response from Get: %+v", body)
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
