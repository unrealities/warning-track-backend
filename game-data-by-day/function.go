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

const (
	databaseCollection = "game-data-by-day"
	dateFormat         = "01-02-2006"
	duration           = 60 * time.Second
	firebaseDomain     = "firebaseio.com"
	logName            = "get-game-data-by-day"
	projectID          = "warning-track-backend"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
//
// ex.: https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {'"date":"03-01-2020"'}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	defer cancel()

	lg, err := gcloud.CloudLogger(ctx, projectID, logName)
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger")
	}

	collection, err := gcloud.FireStoreCollection(ctx, databaseCollection, firebaseDomain, projectID, lg)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gcloud.LogMessage{Message: fmt.Sprintf("error setting up connection to FireStore: %s", err)}})
		return
	}

	date, err := parseDate(r.Body, dateFormat, lg)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gcloud.LogMessage{Message: fmt.Sprintf("error parsing date requested: %s", err)}})
		return
	}

	daySchedule, err := mlbStats.GetSchedule(date, lg)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gcloud.LogMessage{Message: fmt.Sprintf("error getting the daily StatsAPI schedule: %s", err)}})
		return
	}

	// Integrate with FireStore to persist data
	_, err = collection.Doc(date.Format(dateFormat)).Set(ctx, daySchedule)
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
