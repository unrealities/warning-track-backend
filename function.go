package function

import (
	"encoding/json"
	"net/http"

	"github.com/unrealities/warning-track-backend/mlbstats"
	"github.com/unrealities/warning-track-backend/transformers"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"data": {"date":"03-01-2020"}}
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
	daySchedule, err := mlbstats.GetSchedule(s.Date) // Execution Time: ~1000ms
	if err != nil {
		s.HandleFatalError("error getting the daily StatsAPI schedule", err)
	}
	s.DebugMsg("successfully fetched schedule")

	// Transform
	games, err := transformers.OptimusPrime(s.Date, daySchedule)
	if err != nil {
		s.HandleFatalError("error transforming StatsAPI schedule to simpler games struct", err)
	}
	s.DebugMsg("successfully transformed data")

	// Load
	_, err = s.FirestoreClient.Collection(s.DBCollection).Doc(date.Format(s.DateFmt)).Set(ctx, games) // Execution Time: ~ 3500ms
	if err != nil {
		s.HandleFatalError("error persisting data to Firebase", err)
	}

	// CORS
	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusOK)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}
