package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
//
// ex.: https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {'"date":"03-01-20"'}
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

	parsedDate, err := time.Parse("2006-01-02", d.Date)
	if err != nil {
		logger.Printf("Error parsing date requested: %s", err)
	}

	URL := statsAPIScheduledURL(parsedDate)
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
	fmt.Fprintf(w, string(body))
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
