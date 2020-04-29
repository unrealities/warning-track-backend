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
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	"contrib.go.opencensus.io/exporter/stackdriver"
	firebase "firebase.google.com/go"
	"go.opencensus.io/trace"

	"github.com/unrealities/warning-track-backend/mlbStats"
)

// GameDataByDay stores necessary information for the cloud function
type GameDataByDay struct {
	DateFmt         string
	DbCollection    string
	ErrorReporter   *errorreporting.Client
	FirestoreClient *firestore.Client
	FunctionName    string
	Logger          *logging.Client
	ProjectID       string
	TraceSpan       *trace.Span
	Version         string
}

// LogMessage is a simple struct to ensure JSON formatting in logs
type LogMessage struct {
	Err      string `json:",omitempty"`
	Function string `json:"fn"`
	Msg      string `json:"msg"`
	Version  string `json:"version"`
}

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"date":"03-01-2020"}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	gameDataByDay := GameDataByDay{
		DateFmt:      os.Getenv("DATE_FMT"),
		DbCollection: os.Getenv("DB_COLLECTION"),
		ProjectID:    os.Getenv("PROJECT_ID"),
		FunctionName: os.Getenv("FN_NAME"),
		Version:      os.Getenv("VERSION"),
	}
	log.Printf("running version: %s", gameDataByDay.Version)

	// Tracing
	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: gameDataByDay.ProjectID})
	if err != nil {
		log.Fatalf("error setting up OpenCensus Stackdriver Trace exporter: %s", err)
	}
	trace.RegisterExporter(exporter)
	ctx, span := trace.StartSpan(r.Context(), gameDataByDay.FunctionName)
	gameDataByDay.TraceSpan = span

	// Error Reporting
	errorClient, err := errorreporting.NewClient(ctx, gameDataByDay.ProjectID, errorreporting.Config{
		ServiceName:    gameDataByDay.FunctionName,
		ServiceVersion: gameDataByDay.Version,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("error setting up Error Reporting: %s", err)
	}
	defer errorClient.Close()
	gameDataByDay.ErrorReporter = errorClient

	// Cloud Logging
	logClient, err := logging.NewClient(ctx, gameDataByDay.ProjectID)
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger: %s", err)
	}
	defer logClient.Close()
	logClient.OnError = func(e error) {
		fmt.Fprintf(os.Stderr, "logClient error: %v", e)
	}
	gameDataByDay.Logger = logClient

	// Firestore
	conf := &firebase.Config{ProjectID: gameDataByDay.ProjectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		gameDataByDay.HandleFatalError("error setting up Firebase app", err)
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		gameDataByDay.HandleFatalError("error setting up Firestore client", err)
	}
	collection := fsClient.Collection(gameDataByDay.DbCollection)

	gameDataByDay.DebugMsg("successfully fetched Firestore collection")

	date, err := parseDate(r.Body, gameDataByDay.DateFmt)
	if err != nil {
		gameDataByDay.HandleFatalError("error parsing date requested", err)
	}

	daySchedule, err := mlbStats.GetSchedule(date)
	if err != nil {
		gameDataByDay.HandleFatalError("error getting the daily StatsAPI schedule", err)
	}
	gameDataByDay.DebugMsg("successfully fetched schedule")

	_, err = collection.Doc(date.Format(gameDataByDay.DateFmt)).Set(ctx, daySchedule)
	if err != nil {
		gameDataByDay.HandleFatalError("error persisting data to Firebase", err)
	}

	gameDataByDay.DebugMsg("stackdriver debug message success")

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

// HandleFatalError produces an error report, cloud log message and standard log fatal
func (g GameDataByDay) HandleFatalError(msg string, err error) {
	g.ErrorReporter.Report(errorreporting.Entry{Error: err})
	g.Logger.Logger(g.FunctionName).Log(logging.Entry{
		LogName:  g.FunctionName,
		Severity: logging.Error,
		Payload: LogMessage{
			Msg:      msg,
			Err:      err.Error(),
			Function: g.FunctionName,
			Version:  g.Version,
		},
		Trace:  fmt.Sprintf("projects/%s/trace/%s", g.ProjectID, g.TraceSpan.SpanContext().TraceID.String()),
		SpanID: g.TraceSpan.SpanContext().SpanID.String(),
	})
	log.Fatalf("%s: %s", msg, err)
}

// DebugMsg logs a simple debug message with function name and version
func (g GameDataByDay) DebugMsg(msg string) {
	g.Logger.Logger(g.FunctionName).Log(logging.Entry{
		LogName:  g.FunctionName,
		Severity: logging.Debug,
		Payload: LogMessage{
			Msg:      msg,
			Function: g.FunctionName,
			Version:  g.Version,
		},
	})
}
