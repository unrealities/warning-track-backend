package function

import (
	"encoding/json"
	"net/http"

	"github.com/unrealities/warning-track-backend/mlbStats"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"date":"03-01-2020"}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s, err := InitService(ctx) // Execution Time: ~300ms
	if err != nil {
		s.HandleFatalError("error initializing service", err)
	}
	defer s.ErrorReporter.Close()
	defer s.FirestoreClient.Close()
	defer s.Logger.Close()
	defer s.TraceSpan.End()

	date, err := ParseDate(r.Body, s.DateFmt)
	if err != nil {
		s.HandleFatalError("error parsing date requested", err)
	}
	s.Date = date

	// Extract
	daySchedule, err := mlbStats.GetSchedule(s.Date) // Execution Time: ~1000ms
	if err != nil {
		s.HandleFatalError("error getting the daily StatsAPI schedule", err)
	}
	s.DebugMsg("successfully fetched schedule")

	// Transform

	// Load
	_, err = s.FirestoreClient.Collection(s.DBCollection).Doc(date.Format(s.DateFmt)).Set(ctx, daySchedule) // Execution Time: ~ 3500ms
	if err != nil {
		s.HandleFatalError("error persisting data to Firebase", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule)
}
