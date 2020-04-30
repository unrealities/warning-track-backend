package function

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/logging"
	"contrib.go.opencensus.io/exporter/stackdriver"
	firebase "firebase.google.com/go"
	"go.opencensus.io/trace"

	"github.com/unrealities/warning-track-backend/mlbStats"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"date":"03-01-2020"}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	s := Service{
		DateFmt:      os.Getenv("DATE_FMT"),
		DbCollection: os.Getenv("DB_COLLECTION"),
		ProjectID:    os.Getenv("PROJECT_ID"),
		FunctionName: os.Getenv("FN_NAME"),
		Version:      os.Getenv("VERSION"),
	}

	// Tracing
	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: s.ProjectID})
	if err != nil {
		log.Fatalf("error setting up OpenCensus Stackdriver Trace exporter: %s", err)
	}
	trace.RegisterExporter(exporter)
	ctx, span := trace.StartSpan(r.Context(), s.FunctionName)
	s.TraceSpan = span

	// Error Reporting
	errorClient, err := errorreporting.NewClient(ctx, s.ProjectID, errorreporting.Config{
		ServiceName:    s.FunctionName,
		ServiceVersion: s.Version,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("error setting up Error Reporting: %s", err)
	}
	defer errorClient.Close()
	s.ErrorReporter = errorClient

	// Cloud Logging
	logClient, err := logging.NewClient(ctx, s.ProjectID)
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger: %s", err)
	}
	defer logClient.Close()
	logClient.OnError = func(e error) {
		fmt.Fprintf(os.Stderr, "logClient error: %v", e)
	}
	s.Logger = logClient

	// Firestore
	conf := &firebase.Config{ProjectID: s.ProjectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		s.HandleFatalError("error setting up Firebase app", err)
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		s.HandleFatalError("error setting up Firestore client", err)
	}
	collection := fsClient.Collection(s.DbCollection)

	s.DebugMsg("successfully initialized clients")

	date, err := parseDate(r.Body, s.DateFmt)
	if err != nil {
		s.HandleFatalError("error parsing date requested", err)
	}

	daySchedule, err := mlbStats.GetSchedule(date)
	if err != nil {
		s.HandleFatalError("error getting the daily StatsAPI schedule", err)
	}
	s.DebugMsg("successfully fetched schedule")

	_, err = collection.Doc(date.Format(s.DateFmt)).Set(ctx, daySchedule)
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
