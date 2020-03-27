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
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	logger, err := logger()
	if err != nil {
		log.Printf("Error setuping logger")
		return
	}

	logger.Printf("Received request: %+v", r.Body)

	var d struct {
		Date time.Time `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		logger.Printf("Error attempting to decode json body: %s", err)
		return
	}
	logger.Printf("Date requested: %+v", d.Date)

	URL := statsAPIScheduledURL(d.Date)
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
}

// logger returns a logger to create appropriate logs in Google Cloud
func logger() logging.Logger {
	ctx := context.Background
	logName := "get-game-data-by-day"
	projectID := "warning-track-backend"
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.Logger(logName).StandardLogger(logging.Info), nil
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
