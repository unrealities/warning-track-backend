package function

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
//
// ex.: https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {'"date":"03-01-2020"'}
//
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	// setup Logger
	ctx := context.Background()
	logName := "get-game-data-by-day"
	projectID := "warning-track-backend"
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Error setuping logger")
		return
	}
	defer client.Close()
	logger := client.Logger(logName).StandardLogger(logging.Info)

	logger.Printf("Received request: %+v", r.Body)

	var d struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		logger.Printf("Error attempting to decode json body: %s", err)
		return
	}
	logger.Printf("Date requested: %+v", d.Date)

	parsedDate, err := time.Parse("01-02-2006", d.Date)
	if err != nil {
		logger.Printf("Error parsing date requested: %s", err)
	}

	URL := statsAPIScheduleURL(parsedDate)
	logger.Printf("Making Get request: %s", URL)
	resp, err := http.Get(URL)
	if err != nil {
		logger.Printf("Error in Get request: %s", err)
		return
	}
	defer resp.Body.Close()

	logger.Printf("Parsing response from Get request: %+v", resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Error reading Get response body: %s", err)
		return
	}

	logger.Printf("successfully received response from Get: %+v", body)

	statsAPIScheduleResp := statsAPISchedule{}
	err = json.Unmarshal(body, &statsAPIScheduleResp)
	if err != nil {
		logger.Printf("Error trying to unmarshal response from statsAPI: %s", err)
		return
	}

	// TODO: Integrate with Firebase to persist data

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statsAPIScheduleResp)
}

// fireBaseUpdater creates or updates the firebase entry for the given day
func fireBaseUpdater(time time.Time) {
	// TODO: Need to make a create or update
	// https://firebase.google.com/docs/reference/rest/database
}

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
