package function

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/unrealities/warning-track-backend/mlbStats"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"date":"03-01-2020"}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s, err := InitService(ctx)
	if err != nil {
		s.HandleFatalError("error initializing service", err)
	}

	date, err := parseDate(r.Body, s.DateFmt)
	if err != nil {
		s.HandleFatalError("error parsing date requested", err)
	}
	s.Date = date

	daySchedule, err := mlbStats.GetSchedule(s.Date)
	if err != nil {
		s.HandleFatalError("error getting the daily StatsAPI schedule", err)
	}
	s.DebugMsg("successfully fetched schedule")

	_, err = s.FirestoreClient.Collection(s.DbCollection).Doc(date.Format(s.DateFmt)).Set(ctx, daySchedule)
	if err != nil {
		s.HandleFatalError("error persisting data to Firebase", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule)
}

// parseDate parses the request body and returns a time.Time value of the requested date
func parseDate(reqBody io.ReadCloser, dateFormat string) (time.Time, error) {
	var d struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(reqBody).Decode(&d); err != nil {
		return time.Time{}, err
	}

	return time.Parse(dateFormat, d.Date)
}
