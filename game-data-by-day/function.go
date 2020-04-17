package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
	"github.com/unrealities/warning-track-backend/game-data-by-day/gcloud"
	"github.com/unrealities/warning-track-backend/game-data-by-day/mlbStats"
)

// gameDataByDay stores necessary information for the cloud function
type gameDataByDay struct {
	dateFmt        string
	dbCollection   string
	duration       time.Duration
	firebaseDomain string
	logger         *logging.Logger
	projectID      string
}

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
//
// ex.: https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {'"date":"03-01-2020"'}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	gameDataByDay := gameDataByDay{
		dateFmt:        "01-02-2006",
		dbCollection:   "game-data-by-day",
		duration:       60 * time.Second,
		firebaseDomain: "firebaseio.com",
		projectID:      "warning-track-backend",
	}

	ctx, cancel := context.WithTimeout(r.Context(), gameDataByDay.duration)
	defer cancel()

	lg, err := gcloud.CloudLogger(ctx, gameDataByDay.projectID, "get-game-data-by-day")
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger")
	}
	gameDataByDay.logger = lg

	collection, err := gcloud.FireStoreCollection(ctx, gameDataByDay.dbCollection, gameDataByDay.firebaseDomain, gameDataByDay.projectID, gameDataByDay.logger)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gcloud.LogMessage{Message: fmt.Sprintf("error setting up connection to FireStore: %s", err)}})
		return
	}

	date, err := parseDate(r.Body, gameDataByDay.dateFmt, gameDataByDay.logger)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gcloud.LogMessage{Message: fmt.Sprintf("error parsing date requested: %s", err)}})
		return
	}

	daySchedule, err := mlbStats.GetSchedule(date, gameDataByDay.logger)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gcloud.LogMessage{Message: fmt.Sprintf("error getting the daily StatsAPI schedule: %s", err)}})
		return
	}

	// Integrate with FireStore to persist data
	_, err = collection.Doc(date.Format(gameDataByDay.dateFmt)).Set(ctx, daySchedule)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gcloud.LogMessage{Message: fmt.Sprintf("error trying to set value in Firebase: %s", err)}})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule)
}

// parseDate parses the request body and returns a time.Time value of the requested date
func parseDate(reqBody io.ReadCloser, dateFormat string, lg *logging.Logger) (time.Time, error) {
	var d struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(reqBody).Decode(&d); err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gcloud.LogMessage{Message: fmt.Sprintf("error attempting to decode json body: %s", err)}})
		return time.Time{}, err
	}
	lg.Log(logging.Entry{Severity: logging.Debug, Payload: gcloud.LogMessage{Message: fmt.Sprintf("date requested: %+v", d.Date)}})

	return time.Parse(dateFormat, d.Date)
}
