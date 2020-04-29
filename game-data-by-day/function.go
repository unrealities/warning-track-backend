package function

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	"contrib.go.opencensus.io/exporter/stackdriver"
	firebase "firebase.google.com/go"
	"go.opencensus.io/trace"

	"github.com/unrealities/warning-track-backend/mlbStats"
)

// gameDataByDay stores necessary information for the cloud function
type gameDataByDay struct {
	dateFmt         string
	dbCollection    string
	errorReporter   *errorreporting.Client
	firestoreClient *firestore.Client
	functionName    string
	logger          *logging.Client
	projectID       string
	timeout         time.Duration
	version         string
}

// logMessage is a simple struct to ensure JSON formatting in logs
type logMessage struct {
	err      string
	function string
	msg      string
	version  string
}

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"date":"03-01-2020"}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	gameDataByDay := gameDataByDay{
		dateFmt:      "01-02-2006",
		dbCollection: "game-data-by-day",
		projectID:    "warning-track-backend",
		functionName: "GetGameDataByDay",
		timeout:      60 * time.Second,
		version:      "v0.0.52",
	}
	log.Printf("running version: %s", gameDataByDay.version)

	var err error
	ctx := context.Background()

	// Tracing
	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: gameDataByDay.projectID})
	if err != nil {
		log.Fatalf("error setting up OpenCensus Stackdriver Trace exporter: %s", err)
	}
	trace.RegisterExporter(exporter)

	// Error Reporting
	errorClient, err := errorreporting.NewClient(ctx, gameDataByDay.projectID, errorreporting.Config{
		ServiceName:    gameDataByDay.functionName,
		ServiceVersion: gameDataByDay.version,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("error setting up Error Reporting: %s", err)
	}
	defer errorClient.Close()
	gameDataByDay.errorReporter = errorClient

	// Cloud Logging
	logClient, err := logging.NewClient(ctx, gameDataByDay.projectID)
	if err := logClient.Close(); err != nil {
		log.Fatalf("error setting up Google Cloud logger: %s", err)
	}
	defer logClient.Close()
	gameDataByDay.logger = logClient

	gameDataByDay.debugMsg("successfully initialized metrics")

	// Firestore
	conf := &firebase.Config{ProjectID: gameDataByDay.projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		gameDataByDay.handleFatalError("error setting up Firebase app", err)
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		gameDataByDay.handleFatalError("error setting up Firestore client", err)
	}
	collection := fsClient.Collection(gameDataByDay.dbCollection)

	gameDataByDay.debugMsg("successfully fetched Firestore collection")

	date, err := parseDate(r.Body, gameDataByDay.dateFmt)
	if err != nil {
		gameDataByDay.handleFatalError("error parsing date requested", err)
	}

	daySchedule, err := mlbStats.GetSchedule(date)
	if err != nil {
		gameDataByDay.handleFatalError("error getting the daily StatsAPI schedule", err)
	}
	gameDataByDay.debugMsg("successfully fetched schedule")

	log.Printf("daySchedule.Dates[0].Games[0].GameNumber: %v", daySchedule.Dates[0].Games[0].GameNumber)
	_, err = collection.Doc(date.Format(gameDataByDay.dateFmt)).Set(ctx, daySchedule.Dates[0].Games[0])
	if err != nil {
		gameDataByDay.handleFatalError("error persisting data to Firebase", err)
	}

	gameDataByDay.debugMsg("successful run")

	// prevent panic and close logger
	err = gameDataByDay.logger.Logger(gameDataByDay.functionName).Flush()
	if err != nil {
		log.Fatalf("error tring to flush cloud logger: %s", err)
	}
	err = gameDataByDay.logger.Close()
	if err != nil {
		log.Fatalf("error tring to close cloud logger client: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule.Dates[0].Games[0])
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

// handleFatalError produces an error report, cloud log message and standard log fatal
// TODO: include tracing (Trace in logging.Entry)
func (g gameDataByDay) handleFatalError(msg string, err error) {
	g.errorReporter.Report(errorreporting.Entry{Error: err})
	g.logger.Logger(g.functionName).Log(logging.Entry{
		Severity: logging.Error,
		Payload: logMessage{
			msg:      msg,
			err:      err.Error(),
			function: g.functionName,
			version:  g.version,
		},
	})
	log.Fatalf("%s: %s", msg, err)
}

// debugMsg logs a simple debug message with function name and version
func (g gameDataByDay) debugMsg(msg string) {
	g.logger.Logger(g.functionName).Log(logging.Entry{
		Severity: logging.Debug,
		Payload: logMessage{
			msg:      msg,
			function: g.functionName,
			version:  g.version,
		},
	})
	log.Println(msg)
}
