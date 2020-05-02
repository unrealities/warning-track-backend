package function

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/unrealities/warning-track-backend/mlbStats"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"date":"03-01-2020"}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()
	s, err := InitService(ctx)
	if err != nil {
		s.HandleFatalError("error initializing service", err)
	}
	defer s.ErrorReporter.Close()
	defer s.FirestoreClient.Close()
	defer s.Logger.Close()
	defer s.TraceSpan.End()

	log.Printf("InitService: %s", time.Since(start))

	date, err := parseDate(r.Body, s.DateFmt)
	if err != nil {
		s.HandleFatalError("error parsing date requested", err)
	}
	log.Printf("parseDate: %s", time.Since(start))
	s.Date = date
	log.Printf("s.Date: %s", time.Since(start))

	daySchedule, err := mlbStats.GetSchedule(s.Date)
	if err != nil {
		s.HandleFatalError("error getting the daily StatsAPI schedule", err)
	}
	log.Printf("daySchedule: %s", time.Since(start))
	s.DebugMsg("successfully fetched schedule")
	log.Printf("Debug daySchedule: %s", time.Since(start))

	_, err = s.FirestoreClient.Collection(s.DBCollection).Doc(date.Format(s.DateFmt)).Set(ctx, daySchedule)
	if err != nil {
		s.HandleFatalError("error persisting data to Firebase", err)
	}
	log.Printf("Firestore: %s", time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule)

	log.Printf("Finished: %s", time.Since(start))
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
